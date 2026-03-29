package repository

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type storedMetric struct {
	ID    string   `json:"id"`
	Type  string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type MemRepository struct {
	mu sync.RWMutex       // мьютекс для защиты доступа
	gauges   map[string]float64 // метрики типа gauge
	counters map[string]int64   // метрики типа counter

	filePath      string        // файл для сохранения метрик
	storeInterval time.Duration // интервал автосохранения

	saveMu sync.Mutex   // защита записи в файл
	ticker *time.Ticker // тикер для автосохранения
	stopCh chan struct{} // сигнал остановки тикера

	logger *slog.Logger // логирование
}

func NewMemRepository(logger *slog.Logger, filePath string, storeIntervalSeconds int64, restore bool) *MemRepository {
	repo := &MemRepository{
		logger:   logger,                 // логгер
		gauges:   make(map[string]float64), // инициализация хранилища gauge
		counters: make(map[string]int64),   // инициализация хранилища counter
	}

	repo.configureStorage(filePath, storeIntervalSeconds) // настройка файла и интервала сохранения

	if restore {
		// загрузка метрик из файла при старте
		if err := repo.loadFromFile(); err != nil {
			logger.Error("failed to restore metrics", slog.Any("error", err))
		}
	}

	repo.startSaver() // запуск периодического сохранения (если задан интервал)

	return repo
}

func (m *MemRepository) SetGauge(name string, value float64) {
	m.mu.Lock()
	m.gauges[name] = value // сохраняем значение gauge
	m.mu.Unlock()

	// если автосохранение отключено, сохраняем сразу
	if m.storeInterval == 0 {
		if err := m.SaveToFile(); err != nil {
			m.logger.Error("failed to save metrics", slog.Any("error", err))
		}
	}
}

func (m *MemRepository) AddCounter(name string, value int64) {
	m.mu.Lock()
	m.counters[name] += value // увеличиваем значение counter
	m.mu.Unlock()

	// если автосохранение отключено, сохраняем сразу
	if m.storeInterval == 0 {
		if err := m.SaveToFile(); err != nil {
			m.logger.Error("failed to save metrics", slog.Any("error", err))
		}
	}
}

func (m *MemRepository) GetGauge(name string) (float64, bool) {
	m.mu.RLock()         // блокировка на чтение
	defer m.mu.RUnlock() // разблокировка после чтения

	val, ok := m.gauges[name] // получаем значение gauge
	return val, ok            // возвращаем значение и флаг
}

func (m *MemRepository) GetCounter(name string) (int64, bool) {
	m.mu.RLock()         // блокировка на чтение
	defer m.mu.RUnlock() // разблокировка после чтения

	val, ok := m.counters[name] // получаем значение counter
	return val, ok            // возвращаем значение и флаг
}

func (m *MemRepository) GetAll() (map[string]float64, map[string]int64) {
	m.mu.RLock()         // блокировка на чтение
	defer m.mu.RUnlock() // разблокировка после чтения

	// создаём копии, чтобы внешние вызовы не меняли внутренние мапы
	gaugesCopy := make(map[string]float64, len(m.gauges))
	for k, v := range m.gauges {
		gaugesCopy[k] = v
	}

	countersCopy := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		countersCopy[k] = v
	}

	return gaugesCopy, countersCopy
}

func (m *MemRepository) configureStorage(filePath string, storeIntervalSeconds int64) {
	m.filePath = filePath // путь к файлу для сохранения метрик

	if storeIntervalSeconds <= 0 {
		m.storeInterval = 0 // отключаем автосохранение
		return
	}

	m.storeInterval = time.Duration(storeIntervalSeconds) * time.Second // интервал автосохранения
}

func (m *MemRepository) startSaver() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// ничего не делаем, если файл не указан или интервал <= 0
	if m.filePath == "" || m.storeInterval <= 0 {
		return
	}
	// если тикер уже создан, повторно не запускаем
	if m.ticker != nil {
		return
	}

	m.stopCh = make(chan struct{})           // канал для остановки тикера
	m.ticker = time.NewTicker(m.storeInterval) // создаём тикер с заданным интервалом

	go func() {
		for {
			select {
			case <-m.ticker.C:
				_ = m.SaveToFile() // сохраняем метрики при каждом тикe
			case <-m.stopCh:
				m.ticker.Stop() // остановка тикера
				return
			}
		}
	}()
}

func (m *MemRepository) loadFromFile() error {
	if m.filePath == "" {
		return nil // файл не указан — ничего не делаем
	}

	raw, err := os.ReadFile(m.filePath) // читаем файл
	if err != nil {
		if os.IsNotExist(err) {
			return nil // если файла нет — считаем пустым хранилищем
		}
		return err // другие ошибки чтения
	}

	var items []storedMetric
	if err := json.Unmarshal(raw, &items); err != nil {
		return err // ошибка парсинга JSON
	}

	// временные мапы для загрузки
	gauges := make(map[string]float64)
	counters := make(map[string]int64)
	for _, item := range items {
		switch item.Type {
		case "gauge":
			if item.Value != nil {
				gauges[item.ID] = *item.Value
			}
		case "counter":
			if item.Delta != nil {
				counters[item.ID] = *item.Delta
			}
		}
	}

	// подменяем текущие мапы на загруженные
	m.mu.Lock()
	m.gauges = gauges
	m.counters = counters
	m.mu.Unlock()
	return nil
}

func (m *MemRepository) SaveToFile() error {
	if m.filePath == "" {
		return nil // файл не указан — ничего не сохраняем
	}

	m.saveMu.Lock()
	defer m.saveMu.Unlock() // защита записи в файл

	// создаём копии текущих метрик для безопасной записи
	m.mu.RLock()
	gaugesCopy := make(map[string]float64, len(m.gauges))
	for k, v := range m.gauges {
		gaugesCopy[k] = v
	}
	countersCopy := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		countersCopy[k] = v
	}
	m.mu.RUnlock()

	// формируем JSON
	items := make([]storedMetric, 0, len(gaugesCopy)+len(countersCopy))
	for name, val := range gaugesCopy {
		v := val
		items = append(items, storedMetric{
			ID:    name,
			Type:  "gauge",
			Value: &v,
		})
	}
	for name, val := range countersCopy {
		v := val
		items = append(items, storedMetric{
			ID:    name,
			Type:  "counter",
			Delta: &v,
		})
	}

	raw, err := json.Marshal(items) // сериализация в JSON
	if err != nil {
		return err
	}

	// создаём директорию, если нужно
	dir := filepath.Dir(m.filePath)
	base := filepath.Base(m.filePath)
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}

	// запись через временный файл для атомарности
	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	// атомарная замена основного файла
	if err := os.Rename(tmpName, m.filePath); err != nil {
		return err
	}
	return nil
}