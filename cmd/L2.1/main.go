package main

import (
	"fmt"
)

// 1. Паттерн Фасад
// Предоставляет упрощённый интерфейс к сложной подсистеме.

// Компоненты подсистемы
type CPU struct{}

func (c *CPU) Start() { fmt.Println("CPU запущен") }

type Memory struct{}

func (m *Memory) Load() { fmt.Println("Память загружена") }

type Disk struct{}

func (d *Disk) Read() { fmt.Println("Диск прочитан") }

// Фасад
type Computer struct {
	cpu    *CPU
	memory *Memory
	disk   *Disk
}

func NewComputer() *Computer {
	return &Computer{
		cpu:    &CPU{},
		memory: &Memory{},
		disk:   &Disk{},
	}
}

func (c *Computer) Start() {
	c.cpu.Start()
	c.memory.Load()
	c.disk.Read()
}

// Плюсы: Упрощает работу со сложными подсистемами, улучшает удобство использования.
// Минусы: Снижает прозрачность, может ввести избыточность.
// Пример: Запуск компьютера включает множество подсистем (CPU, память, диск), фасад упрощает этот процесс.

// 2. Паттерн Строитель
// Постепенное создание сложного объекта.

type House struct {
	walls   string
	roof    string
	windows int
}

type HouseBuilder struct {
	house *House
}

func NewHouseBuilder() *HouseBuilder {
	return &HouseBuilder{house: &House{}}
}

func (b *HouseBuilder) BuildWalls(walls string) *HouseBuilder {
	b.house.walls = walls
	return b
}

func (b *HouseBuilder) BuildRoof(roof string) *HouseBuilder {
	b.house.roof = roof
	return b
}

func (b *HouseBuilder) BuildWindows(windows int) *HouseBuilder {
	b.house.windows = windows
	return b
}

func (b *HouseBuilder) GetResult() *House {
	return b.house
}

// Плюсы: Гибкость в создании объектов, разделение процесса создания и представления.
// Минусы: Увеличивает сложность, может не подходить для простых объектов.
// Пример: Создание настраиваемых домов с различными характеристиками.

// 3. Паттерн Посетитель
// Позволяет добавлять новые операции к объектам без изменения их структуры.

type Element interface {
	Accept(Visitor)
}

type ConcreteElementA struct{}

func (e *ConcreteElementA) Accept(v Visitor) { v.VisitConcreteElementA(e) }

type ConcreteElementB struct{}

func (e *ConcreteElementB) Accept(v Visitor) { v.VisitConcreteElementB(e) }

// Интерфейс Посетителя
type Visitor interface {
	VisitConcreteElementA(*ConcreteElementA)
	VisitConcreteElementB(*ConcreteElementB)
}

type ConcreteVisitor struct{}

func (v *ConcreteVisitor) VisitConcreteElementA(e *ConcreteElementA) {
	fmt.Println("Посетили ConcreteElementA")
}
func (v *ConcreteVisitor) VisitConcreteElementB(e *ConcreteElementB) {
	fmt.Println("Посетили ConcreteElementB")
}

// Плюсы: Добавление операций без изменения элементов.
// Минусы: Сложность для больших иерархий.
// Пример: Расчёт налогов для различных типов продуктов.

// 4. Паттерн Команда
// Инкапсулирует запрос как объект, позволяя параметризовать и ставить в очередь.

type Command interface {
	Execute()
}

type Light struct{}

func (l *Light) On()  { fmt.Println("Свет включён") }
func (l *Light) Off() { fmt.Println("Свет выключен") }

type LightOnCommand struct {
	light *Light
}

func (c *LightOnCommand) Execute() { c.light.On() }

type LightOffCommand struct {
	light *Light
}

func (c *LightOffCommand) Execute() { c.light.Off() }

// Плюсы: Разделение отправителя и получателя, поддержка отмены/повтора действий.
// Минусы: Может привести к созданию большого числа классов.
// Пример: Системы автоматизации дома.

// 5. Цепочка обязанностей
// Передача запросов по цепочке обработчиков.

type Handler interface {
	SetNext(Handler) Handler
	HandleRequest(request string)
}

type BaseHandler struct {
	next Handler
}

func (h *BaseHandler) SetNext(next Handler) Handler {
	h.next = next
	return next
}

func (h *BaseHandler) HandleRequest(request string) {
	if h.next != nil {
		h.next.HandleRequest(request)
	}
}

type ConcreteHandlerA struct {
	BaseHandler
}

func (h *ConcreteHandlerA) HandleRequest(request string) {
	if request == "A" {
		fmt.Println("HandlerA обработал запрос")
	} else {
		h.BaseHandler.HandleRequest(request)
	}
}

// Плюсы: Гибкость конфигурации цепочек, способствует слабой связанности.
// Минусы: Нет гарантии, что запрос будет обработан.
// Пример: Эскалация тикетов в службе поддержки.

// 6. Фабричный метод
// Предоставляет интерфейс для создания объектов, позволяя подклассам изменять тип создаваемых объектов.

type Product interface {
	Use() string
}

type ConcreteProductA struct{}

func (p *ConcreteProductA) Use() string { return "Использование Продукта A" }

type ConcreteProductB struct{}

func (p *ConcreteProductB) Use() string { return "Использование Продукта B" }

type Creator interface {
	CreateProduct() Product
}

type ConcreteCreatorA struct{}

func (c *ConcreteCreatorA) CreateProduct() Product { return &ConcreteProductA{} }

type ConcreteCreatorB struct{}

func (c *ConcreteCreatorB) CreateProduct() Product { return &ConcreteProductB{} }

// Плюсы: Содействует консистентности, снижает дублирование кода.
// Минусы: Увеличивает сложность, может привести к избыточному числу классов.
// Пример: Создание различных типов документов в редакторе.

// 7. Паттерн Стратегия
// Определяет семейство алгоритмов, инкапсулирует каждый из них и делает их взаимозаменяемыми.

type Strategy interface {
	Execute(a, b int) int
}

type AddStrategy struct{}

func (s *AddStrategy) Execute(a, b int) int { return a + b }

type MultiplyStrategy struct{}

func (s *MultiplyStrategy) Execute(a, b int) int { return a * b }

type Context struct {
	strategy Strategy
}

func (c *Context) SetStrategy(s Strategy) { c.strategy = s }
func (c *Context) ExecuteStrategy(a, b int) int {
	return c.strategy.Execute(a, b)
}

// Плюсы: Лёгкая смена поведения, поддержка принципа открытости/закрытости.
// Минусы: Увеличение количества классов.
// Пример: Стратегии оплаты (кредитная карта, PayPal и т. д.).

// 8. Паттерн Состояние
// Позволяет объекту изменять своё поведение при изменении внутреннего состояния.

type State interface {
	DoAction(context *ContextState)
}

type ContextState struct {
	state State
}

func (c *ContextState) SetState(state State) { c.state = state }
func (c *ContextState) Request() {
	c.state.DoAction(c)
}

type StartState struct{}

func (s *StartState) DoAction(context *ContextState) {
	fmt.Println("Состояние: Запуск")
	context.SetState(&StopState{})
}

type StopState struct{}

func (s *StopState) DoAction(context *ContextState) {
	fmt.Println("Состояние: Останов")
	context.SetState(&StartState{})
}

// Плюсы: Упрощает переходы между состояниями, способствует инкапсуляции.
// Минусы: Может привести к большому числу классов состояний.
// Пример: Состояния медиаплеера (воспроизведение, пауза, остановка).
