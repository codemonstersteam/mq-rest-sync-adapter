package parser

// AsyncAPISpec представляет полную спецификацию AsyncAPI 3.0
type AsyncAPISpec struct {
	AsyncAPI string                       `yaml:"asyncapi"`
	Info     Info                         `yaml:"info"`
	Servers  map[string]Server            `yaml:"servers"`
	Channels map[string]Channel           `yaml:"channels"`
	Operations map[string]Operation       `yaml:"operations"`
	Components *Components                `yaml:"components,omitempty"`
}

// Info содержит метаданные о спецификации
type Info struct {
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
	Description string `yaml:"description,omitempty"`
}

// Server описывает сервер для подключения
type Server struct {
	Host        string            `yaml:"host"`
	Protocol    string            `yaml:"protocol"`
	Description string            `yaml:"description,omitempty"`
	Variables   map[string]interface{} `yaml:"variables,omitempty"`
	Bindings    map[string]interface{} `yaml:"bindings,omitempty"`
}

// Channel представляет канал для обмена сообщениями
type Channel struct {
	Address     string                 `yaml:"address,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Servers     []ServerRef            `yaml:"servers,omitempty"`
	Messages    map[string]MessageRef  `yaml:"messages,omitempty"`
	Bindings    map[string]interface{} `yaml:"bindings,omitempty"`
}

// Operation описывает операцию отправки или получения
type Operation struct {
	Action      string                 `yaml:"action"` // send или receive
	Channel     ChannelRef             `yaml:"channel"`
	Summary     string                 `yaml:"summary,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Messages    []MessageRef           `yaml:"messages,omitempty"`
	Reply       *Reply                 `yaml:"reply,omitempty"`
	Bindings    map[string]interface{} `yaml:"bindings,omitempty"`
}

// Reply описывает ответ на операцию
type Reply struct {
	Channel  ChannelRef   `yaml:"channel,omitempty"`
	Messages []MessageRef `yaml:"messages,omitempty"`
}

// Message представляет структуру сообщения
type Message struct {
	Name        string                 `yaml:"name,omitempty"`
	Title       string                 `yaml:"title,omitempty"`
	Summary     string                 `yaml:"summary,omitempty"`
	ContentType string                 `yaml:"contentType,omitempty"`
	Headers     *Schema                `yaml:"headers,omitempty"`
	Payload     *Schema                `yaml:"payload,omitempty"`
	CorrelationId *CorrelationId       `yaml:"correlationId,omitempty"`
	Bindings    map[string]interface{} `yaml:"bindings,omitempty"`
}

// Schema представляет JSON Schema для payload или headers
type Schema struct {
	Type        string                 `yaml:"type"`                // Обязательное поле - тип всегда должен быть указан
	Properties  map[string]Schema      `yaml:"properties,omitempty"`
	Required    []string               `yaml:"required,omitempty"`
	Example     interface{}            `yaml:"example"`             // Обязательное поле - пример значения
	Format      string                 `yaml:"format,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Enum        []interface{}          `yaml:"enum,omitempty"`
	Items       *Schema                `yaml:"items,omitempty"`
	Ref         string                 `yaml:"$ref,omitempty"`
	Additional  map[string]interface{} `yaml:",inline"` // для дополнительных полей
}

// CorrelationId описывает схему корреляционного идентификатора
type CorrelationId struct {
	Location    string `yaml:"location"`
	Description string `yaml:"description,omitempty"`
}

// Components содержит переиспользуемые компоненты
type Components struct {
	Messages   map[string]Message   `yaml:"messages,omitempty"`
	Schemas    map[string]Schema    `yaml:"schemas,omitempty"`
	Servers    map[string]Server    `yaml:"servers,omitempty"`
	Channels   map[string]Channel   `yaml:"channels,omitempty"`
	Operations map[string]Operation `yaml:"operations,omitempty"`
}

// Reference типы для ссылок
type ServerRef struct {
	Ref string `yaml:"$ref,omitempty"`
}

type ChannelRef struct {
	Ref string `yaml:"$ref,omitempty"`
}

type MessageRef struct {
	// Reference to a message component
	Ref string `yaml:"$ref,omitempty"`
	
	// Inline message definition (embedded Message)
	*Message `yaml:",inline"`
}