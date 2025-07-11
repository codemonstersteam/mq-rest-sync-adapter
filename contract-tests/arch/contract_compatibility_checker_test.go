package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestContractCompatibilityChecker_parseYAMLToJSON(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		wantError bool
	}{
		{
			name: "valid YAML",
			yamlInput: `asyncapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
channels:
  test:
    address: "/api/test"`,
			wantError: false,
		},
		{
			name:      "invalid YAML",
			yamlInput: "invalid: yaml: content: [",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &ContractCompatibilityChecker{}
			result, err := checker.parseYAMLToJSON([]byte(tt.yamlInput))

			if tt.wantError {
				if err == nil {
					t.Errorf("parseYAMLToJSON() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseYAMLToJSON() unexpected error = %v", err)
				return
			}

			if !result.Get("asyncapi").Exists() {
				t.Errorf("parseYAMLToJSON() result missing asyncapi field")
			}
		})
	}
}

func TestContractCompatibilityChecker_normalizeResourcePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "clientId parameter",
			path:     "/api/client/{clientId}/wallet/balance",
			expected: "/api/client/{param}/wallet/balance",
		},
		{
			name:     "Api-clientId parameter",
			path:     "/api/client/{Api-clientId}/wallet/balance",
			expected: "/api/client/{param}/wallet/balance",
		},
		{
			name:     "no parameters",
			path:     "/api/wallet/balance",
			expected: "/api/wallet/balance",
		},
		{
			name:     "multiple parameters",
			path:     "/api/client/{clientId}/wallet/{Api-clientId}/balance",
			expected: "/api/client/{param}/wallet/{param}/balance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &ContractCompatibilityChecker{}
			result := checker.normalizeResourcePath(tt.path)
			if result != tt.expected {
				t.Errorf("normalizeResourcePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContractCompatibilityChecker_areTypesCompatible(t *testing.T) {
	tests := []struct {
		name         string
		consumerType string
		providerType string
		expected     bool
	}{
		{
			name:         "identical types",
			consumerType: "string",
			providerType: "string",
			expected:     true,
		},
		{
			name:         "integer to number",
			consumerType: "integer",
			providerType: "number",
			expected:     true,
		},
		{
			name:         "number to integer",
			consumerType: "number",
			providerType: "integer",
			expected:     true,
		},
		{
			name:         "incompatible types",
			consumerType: "string",
			providerType: "integer",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &ContractCompatibilityChecker{}
			result := checker.areTypesCompatible(tt.consumerType, tt.providerType)
			if result != tt.expected {
				t.Errorf("areTypesCompatible() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContractCompatibilityChecker_inferMessageType(t *testing.T) {
	tests := []struct {
		name        string
		messageName string
		expected    string
	}{
		{
			name:        "request message",
			messageName: "getBalanceRequest",
			expected:    "request",
		},
		{
			name:        "response message",
			messageName: "getBalanceResponse",
			expected:    "response",
		},
		{
			name:        "mixed case request",
			messageName: "HttpBalanceRequest",
			expected:    "request",
		},
		{
			name:        "unknown message type",
			messageName: "someMessage",
			expected:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &ContractCompatibilityChecker{}
			result := checker.inferMessageType(tt.messageName)
			if result != tt.expected {
				t.Errorf("inferMessageType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContractCompatibilityChecker_extractMessageName(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected string
	}{
		{
			name:     "valid message reference",
			ref:      "#/components/messages/getBalanceRequest",
			expected: "getBalanceRequest",
		},
		{
			name:     "invalid reference",
			ref:      "invalid/reference",
			expected: "",
		},
		{
			name:     "empty reference",
			ref:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &ContractCompatibilityChecker{}
			result := checker.extractMessageName(tt.ref)
			if result != tt.expected {
				t.Errorf("extractMessageName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContractCompatibilityChecker_extractSchemaName(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected string
	}{
		{
			name:     "valid schema reference",
			ref:      "#/components/schemas/balanceRequestPayload",
			expected: "balanceRequestPayload",
		},
		{
			name:     "invalid reference",
			ref:      "invalid/reference",
			expected: "",
		},
		{
			name:     "empty reference",
			ref:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &ContractCompatibilityChecker{}
			result := checker.extractSchemaName(tt.ref)
			if result != tt.expected {
				t.Errorf("extractSchemaName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContractCompatibilityChecker_CheckCompatibility_WithTestData(t *testing.T) {
	// Setup test data files
	setupTestData(t)
	defer cleanupTestData()

	tests := []struct {
		name                string
		consumerSpecPath    string
		providerSpecPath    string
		expectedCompatible  bool
		expectedBreakingChanges int
		expectedWarnings    int
		shouldContainError  string
	}{
		{
			name:                "compatible specifications",
			consumerSpecPath:    "testdata/consumer_spec.yml",
			providerSpecPath:    "testdata/provider_spec_valid.yml",
			expectedCompatible:  false, // Still expect some differences
			expectedBreakingChanges: 1, // At least one breaking change expected
			expectedWarnings:    4,     // MQ channel warnings
			shouldContainError:  "",
		},
		{
			name:                "incompatible specifications",
			consumerSpecPath:    "testdata/consumer_spec.yml",
			providerSpecPath:    "testdata/provider_spec_invalid.yml",
			expectedCompatible:  false,
			expectedBreakingChanges: 1, // At least one breaking change expected
			expectedWarnings:    4,     // MQ channel warnings
			shouldContainError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewTestContractCompatibilityChecker(tt.consumerSpecPath, tt.providerSpecPath)
			
			result, err := checker.CheckCompatibility()
			
			if tt.shouldContainError != "" {
				if err == nil || err.Error() != tt.shouldContainError {
					t.Errorf("CheckCompatibility() expected error containing %v, got %v", tt.shouldContainError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("CheckCompatibility() unexpected error = %v", err)
				return
			}

			if result.IsCompatible != tt.expectedCompatible {
				t.Errorf("CheckCompatibility() IsCompatible = %v, want %v", result.IsCompatible, tt.expectedCompatible)
			}

			if len(result.BreakingChanges) < tt.expectedBreakingChanges {
				t.Errorf("CheckCompatibility() BreakingChanges count = %v, want at least %v", len(result.BreakingChanges), tt.expectedBreakingChanges)
			}

			if len(result.Warnings) < tt.expectedWarnings {
				t.Errorf("CheckCompatibility() Warnings count = %v, want at least %v", len(result.Warnings), tt.expectedWarnings)
			}

			// Verify JSON structure
			jsonResult, err := json.Marshal(result)
			if err != nil {
				t.Errorf("Failed to marshal result to JSON: %v", err)
			}

			var parsedResult CompatibilityResult
			if err := json.Unmarshal(jsonResult, &parsedResult); err != nil {
				t.Errorf("Failed to unmarshal result from JSON: %v", err)
			}
		})
	}
}

func TestContractCompatibilityChecker_LoadSpecifications(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData()

	tests := []struct {
		name          string
		specPath      string
		wantError     bool
		shouldContain string
	}{
		{
			name:          "valid consumer spec",
			specPath:      "testdata/consumer_spec.yml",
			wantError:     false,
			shouldContain: "MQ Adapter for REST wallet balance service",
		},
		{
			name:      "non-existent file",
			specPath:  "testdata/non_existent.yml",
			wantError: true,
		},
		{
			name:          "valid provider spec",
			specPath:      "testdata/provider_spec_valid.yml",
			wantError:     false,
			shouldContain: "Wallet Balance Service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &TestContractCompatibilityChecker{
				ConsumerSpecPath: tt.specPath,
			}

			result, err := checker.loadLocalSpec(tt.specPath)

			if tt.wantError {
				if err == nil {
					t.Errorf("loadConsumerSpec() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("loadConsumerSpec() unexpected error = %v", err)
				return
			}

			title := result.Get("info.title").String()
			if tt.shouldContain != "" && title != tt.shouldContain {
				t.Errorf("loadConsumerSpec() title = %v, should contain %v", title, tt.shouldContain)
			}
		})
	}
}

func setupTestData(t *testing.T) {
	// Create testdata directory
	if err := os.MkdirAll("testdata", 0755); err != nil {
		t.Fatalf("Failed to create testdata directory: %v", err)
	}

	// Consumer specification (copy from current consumer spec)
	consumerSpec := `asyncapi: 3.0.0
info:
  title: MQ Adapter for REST wallet balance service
  version: 1.0.0
  description: |
    Пример MQ адаптера для REST-сервиса
    Имеем компонент в системе, который представляет собой синхронный REST сервис.
    Сервис предоставляет информацию об электро кошельке клиента.
    Нам необходимо обеспечить асинхронную работу с данным сервисом,
    для этого мы реализовали адаптер, который с работает с броекром сообщений и передает запрос в REST сервис.
servers:
  broker:
    host: ibmmq://localhost:1414
    protocol: ibmmq
    pathname: /QM1/DEV.ADMIN.SVRCONN
    description: |
      Тестовый инстанс ibm mq в докере
  restServer:
    host: https://localhost:8080
    protocol: https
    pathname: '/api/client/{clientId}/wallet/balance'
    description: |
      Тестовый инстанс REST http в докере 
      который предоставляет информацию о балансе
channels:
  incomingMessages:
    address: "IN.QUEUE.PS"
    description: |
      тригерное сообщение, которое приходит из очереди "запросов"
      канал используется для вычитывания входящих сообщений из очереди "запросов"
    messages:
      httpRequestEnvelope:
        $ref: '#/components/messages/httpRequestEnvelope'
  receiptReceived:
    address: "RECEIPT.QUEUE.PS"
    description: |
      канал используется для доставки квитанции об успешном получении запроса и начале обработке запроса
    messages:
      receiptValidationAcceptance:
        $ref: '#/components/messages/receiptValidationAcceptance'
  transactionReceipt:
    address: "RECEIPT.QUEUE.PS"
    description: канал используется для доставки квитанции о транзакции
    messages:
      transactionReceipt:
        $ref: '#/components/messages/transactionReceipt'
  sendRestResponse:
    address: "OUT.QUEUE.PS"
    description: канал используется для доставки ответа от REST сервиса
    messages:
      httpResponseEnvelope:
        $ref: '#/components/messages/httpResponseEnvelope'
  restGetBalanceRequest:
    address: "/api/client/{Api-clientId}/wallet/balance"
    description: |
      Канал используется для обращения по REST-у в сервис
      для получения баланса кошелька
    messages:
      httpBalanceRequest:
        $ref: '#/components/messages/psGetBalanceRequest'
      httpBalanceResponse:
        $ref: '#/components/messages/psGetBalanceResponse'
operations:
  receiveEnvelope:
    action: receive
    channel:
      $ref: "#/channels/incomingMessages"
    summary: |
      :Получает конверт с запросом
  sendAcceptedReceipt:
    action: send
    channel:
      $ref: "#/channels/receiptReceived"
  sendTransactionReceipt:
    action: send
    channel:
      $ref: "#/channels/transactionReceipt"
  sendRestResponse:
    action: send
    channel:
      $ref: "#/channels/sendRestResponse"
  getBalanceRequest:
    action: send
    channel:
      $ref: "#/channels/restGetBalanceRequest"
    reply:
      channel:
        $ref: "#/channels/restGetBalanceRequest"
      messages:
        - $ref: "#/components/messages/psGetBalanceResponse"
components:
  securitySchemes:
    userPass:
      type: userPassword
  messages:
    httpRequestEnvelope:
      name: httpRequestEnvelope
      title: Http Request Envelope
      contentType: application/json
      payload:
        $ref: '#/components/schemas/httpEnvelopePayload'
    receiptValidationAcceptance:
      name: ReceiptValidationAcceptance
      title: Квитанция об успешном получении запроса
      contentType: application/json
      payload:
        $ref: '#/components/schemas/receiptValidationAcceptancePayload'
    transactionReceipt:
      name: TransactionReceipt
      title: Квитанция об успешном получении запроса
      contentType: application/json
      payload:
        $ref: '#/components/schemas/transactionReceiptPayload'
    httpResponseEnvelope:
      name: httpResponseEnvelope
      title: Http Response Envelope
      contentType: application/json
      payload:
        $ref: '#/components/schemas/httpResponseEnvelopePayload'
    psGetBalanceRequest:
      name: PsGetBalanceRequest
      title: Http Request PsGetBalanceRequest
      contentType: application/json
      headers:
        $ref: '#/components/schemas/psGetBalanceRequestHeaders'
      payload:
        $ref: '#/components/schemas/psGetBalanceRequestPayload'
    psGetBalanceResponse:
      name: PsGetBalanceResponse  
      title: Ответ с балансом кошелька
      contentType: application/json
      headers:
        $ref: '#/components/schemas/responseHeaders'
      payload:
        $ref: '#/components/schemas/balanceResponsePayload'
  schemas:
    httpEnvelopePayload:
      type: object
      required:
        - correlationId
        - method
        - service
        - operation
        - headers
        - body
      properties:
        correlationId:
          type: string
          description: Correlation Id
        method:
          type: string
          enum:
            - GET
            - POST
          description: http метод
        service:
          type: string
          enum:
            - api-gateway
          description: конфигурация вызова в перекладчике содержит название REST сервиса
        operation:
          type: string
          enum:
            - client-wallet-balance
            - client-wallet-status
          description: |
            конфигурация вызова в перекладчике содержит название операции REST сервиса
            из двух свойств service + operation из конфига собирается
            url запроса в REST сервис
        headers:
          type: object
          properties:
            "Api-clientId":
              type: string
              description: |
                Идентификатор клиента содержится в заголовке http запроса
                Этот параметр также используется в формировании запроса в REST сервис
          required:
            - "Api-clientId"
        body:
          type: string
          description: request body
    receiptValidationAcceptancePayload:
      type: object
      required:
        - requestType
        - dateFrom
        - operTime
        - flag
      properties:
        requestType:
          type: string
          description: Тип запроса
          enum:
            - operationHistory
        dateFrom:
          type: number
          description: |
            Дата сформированного запроса 
            Java OffsetDateTime
            A date-time with an offset from UTC/Greenwich in the ISO-8601 calendar system, such as 2007-12-03T10:15:30+01:00.
        operTime:
          type: number
          description: |
            Дата и время операции
            Java OffsetDateTime
            A date-time with an offset from UTC/Greenwich in the ISO-8601 calendar system, such as 2007-12-03T10:15:30+01:00.
        flag:
          type: string
          description: |
            Флаг определяющий статус квитанции
          enum:
            - ACCEPTED
        dateTo:
          type: number
          description: |
            Дата и время окончания запроса
            Java OffsetDateTime
            A date-time with an offset from UTC/Greenwich in the ISO-8601 calendar system, such as 2007-12-03T10:15:30+01:00.
        errorCode:
          type: number
          description: Код ошибки
        description:
          type: string
          description: Описание ошибки
    transactionReceiptPayload:
      type: object
      required:
        - requestType
        - dateFrom
        - operTime
        - flag
      properties:
        requestType:
          type: string
          description: Тип запроса
          enum:
            - operationHistory
        dateFrom:
          type: number
          description: |
            Дата сформированного запроса 
            Java OffsetDateTime
            A date-time with an offset from UTC/Greenwich in the ISO-8601 calendar system, such as 2007-12-03T10:15:30+01:00.
        operTime:
          type: number
          description: |
            Дата и время операции
            Java OffsetDateTime
            A date-time with an offset from UTC/Greenwich in the ISO-8601 calendar system, such as 2007-12-03T10:15:30+01:00.
        flag:
          type: string
          description: |
            Флаг определяющий статус квитанции
          enum:
            - SUCCESS
            - ERROR
        dateTo:
          type: number
          description: |
            Дата и время окончания запроса
            Java OffsetDateTime
            A date-time with an offset from UTC/Greenwich in the ISO-8601 calendar system, such as 2007-12-03T10:15:30+01:00.
        errorCode:
          type: number
          description: Код ошибки
        description:
          type: string
          description: Описание ошибки
    httpResponseEnvelopePayload:
      type: object
      required:
        - correlationId
        - status
        - message
        - method
        - service
        - operation
        - headers
        - body
      properties:
        correlationId:
          type: string
          description: Correlation Id
        status:
          type: integer
          description: httpCode
        message:
          type: string
          description: httpCode.reasonPhrase
        method:
          type: string
          enum:
            - GET
            - POST
          description: http метод
        service:
          type: string
          enum:
            - api-gateway
          description: конфигурация вызова в перекладчике содержит название REST сервиса
        operation:
          type: string
          enum:
            - client-wallet-balance
            - client-wallet-status
          description: |
            конфигурация вызова в перекладчике содержит название операции REST сервиса
            из двух свойств service + operation из конфига собирается
            url запроса в REST сервис
        headers:
          type: object
          properties:
            "Api-clientId":
              type: string
              description: |
                Идентификатор клиента содержится в заголовке http запроса
                Этот параметр также используется в формировании запроса в REST сервис
          required:
            - "Api-clientId"
        body:
          type: string
          description: request body
    psGetBalanceRequestHeaders:
      type: object
      properties:
        Api-clientId:
          type: string
          example: Api-clientId-01
        RequestID:
          type: string
          example: request-01
    psGetBalanceRequestPayload:
      type: object
      required:
        - requestType
        - dateFrom
        - dateTo
        - operTime
      properties:
        requestType:
          type: string
          example: "operationHistory"
        dateFrom:
          type: number
          format: double
          example: 1747377682.865268036
        dateTo:
          type: number
          format: double
          example: 1748241682.865268036
        operTime:
          type: number
          format: double
          example: 1748241672.865268036
    responseHeaders:
      type: object
      properties:
        Content-Type:
          type: string
          enum: ["application/json"]
          example: "application/json"
        RequestID:
          type: string
          description: Идентификатор запроса для корреляции
          example: "request-01"
      required:
        - Content-Type   
    balanceResponsePayload:
      type: object
      required:
        - status
        - actualTimestamp
        - data
      properties:
        status:
          type: string
          description: Статус выполнения запроса
          enum:
            - success
            - error
          example: "success"
        actualTimestamp:
          type: integer
          format: int64
          description: |
            Актуальное время ответа в миллисекундах Unix timestamp
          example: 1607449018829
        data:
          $ref: '#/components/schemas/walletBalanceData'
        error:
          $ref: '#/components/schemas/errorInfo'
    walletBalanceData:
      type: object
      required:
        - createTime
        - operationId
        - balance
      properties:
        createTime:
          type: integer
          format: int64
          description: |
            Время создания записи о балансе в миллисекундах Unix timestamp
          example: 1600443296227
        operationId:
          type: string
          description: Уникальный идентификатор операции
          example: "01"
        balance:
          $ref: '#/components/schemas/balanceInfo'
    balanceInfo:
      type: object
      required:
        - value
        - currency
      properties:
        value:
          type: number
          format: double
          description: Сумма баланса
          example: 999.9
        currency:
          type: string
          description: Код валюты баланса
          enum:
            - rub
            - usd
            - eur
          example: "rub"
    errorInfo:
      type: object
      properties:
        code:
          type: integer
          description: Код ошибки
          example: 1001
        message:
          type: string
          description: Описание ошибки
          example: "Client not found"
        details:
          type: string
          description: Детальная информация об ошибке
          example: "Client with ID Api-clientId-01 does not exist"
`

	// Valid provider specification (main branch)
	providerSpecValid := `asyncapi: 3.0.0
info:
  title: Wallet Balance Service
  version: 1.0.0
  description: |
    Сервис для получения баланса электронного кошелька клиента.
    Сервис принимает HTTP POST запросы и возвращает информацию о балансе
    в формате JSON. Интегрируется с MQ адаптером через REST API.

servers:
  production:
    host: localhost:8080
    protocol: https
    description: Production server для wallet balance service
  development:
    host: localhost:8080  
    protocol: http
    description: Development server для тестирования

channels:
  walletBalanceRequest:
    address: "/api/client/{clientId}/wallet/balance"
    description: |
      Канал для получения запросов на баланс кошелька клиента.
      Принимает POST запросы с данными о периоде и типе операции.
      clientId передается в URL path.
    messages:
      getBalanceRequest:
        $ref: '#/components/messages/getBalanceRequest'
  
  walletBalanceResponse:
    address: "/api/client/{clientId}/wallet/balance"
    description: |
      Канал для отправки ответов с информацией о балансе кошелька.
      clientId передается в URL path.
    messages:
      getBalanceResponse:
        $ref: '#/components/messages/getBalanceResponse'

operations:
  receiveBalanceRequest:
    action: receive
    channel:
      $ref: "#/channels/walletBalanceRequest"
    summary: Получение запроса на баланс кошелька
    description: |
      Операция для получения POST запроса с параметрами
      для получения баланса кошелька клиента
    
  sendBalanceResponse:
    action: send
    channel:
      $ref: "#/channels/walletBalanceResponse"
    summary: Отправка ответа с балансом кошелька
    description: |
      Операция для отправки ответа с информацией о балансе
      кошелька клиента и метаданными операции

components:
  messages:
    getBalanceRequest:
      name: GetBalanceRequest
      title: Запрос на получение баланса кошелька
      contentType: application/json
      headers:
        $ref: '#/components/schemas/requestHeaders'
      payload:
        $ref: '#/components/schemas/balanceRequestPayload'
    
    getBalanceResponse:
      name: GetBalanceResponse  
      title: Ответ с балансом кошелька
      contentType: application/json
      headers:
        $ref: '#/components/schemas/responseHeaders'
      payload:
        $ref: '#/components/schemas/balanceResponsePayload'

  schemas:
    requestHeaders:
      type: object
      properties:
        Api-clientId:
          type: string
          description: Идентификатор клиента из URL path
          example: "Api-clientId-01"
        RequestID:
          type: string
          description: Уникальный идентификатор запроса
          example: "request-01"
        Content-Type:
          type: string
          enum: ["application/json"]
          example: "application/json"
      required:
        - Api-clientId
        - Content-Type

    responseHeaders:
      type: object
      properties:
        Content-Type:
          type: string
          enum: ["application/json"]
          example: "application/json"
        RequestID:
          type: string
          description: Идентификатор запроса для корреляции
          example: "request-01"
      required:
        - Content-Type

    balanceRequestPayload:
      type: object
      required:
        - requestType
        - dateFrom
        - dateTo
        - operTime
      properties:
        requestType:
          type: string
          description: Тип запроса для получения истории операций
          enum:
            - operationHistory
          example: "operationHistory"
        dateFrom:
          type: number
          format: double
          description: |
            Дата начала периода в формате Unix timestamp с микросекундами
            Соответствует Java OffsetDateTime
          example: 1745299787.198193738
        dateTo:
          type: number
          format: double
          description: |
            Дата окончания периода в формате Unix timestamp с микросекундами
            Соответствует Java OffsetDateTime
          example: 1745299792.198546764
        operTime:
          type: number
          format: double
          description: |
            Время операции в формате Unix timestamp с микросекундами
            Соответствует Java OffsetDateTime  
          example: 1745299737.198562113

    balanceResponsePayload:
      type: object
      required:
        - status
        - actualTimestamp
        - data
      properties:
        status:
          type: string
          description: Статус выполнения запроса
          enum:
            - success
            - error
          example: "success"
        actualTimestamp:
          type: integer
          format: int64
          description: |
            Актуальное время ответа в миллисекундах Unix timestamp
          example: 1607449018829
        data:
          $ref: '#/components/schemas/walletBalanceData'
        error:
          $ref: '#/components/schemas/errorInfo'

    walletBalanceData:
      type: object
      required:
        - createTime
        - operationId
        - balance
      properties:
        createTime:
          type: integer
          format: int64
          description: |
            Время создания записи о балансе в миллисекундах Unix timestamp
          example: 1600443296227
        operationId:
          type: string
          description: Уникальный идентификатор операции
          example: "01"
        balance:
          $ref: '#/components/schemas/balanceInfo'

    balanceInfo:
      type: object
      required:
        - value
        - currency
      properties:
        value:
          type: number
          format: double
          description: Сумма баланса
          example: 999.9
        currency:
          type: string
          description: Код валюты баланса
          enum:
            - rub
            - usd
            - eur
          example: "rub"

    errorInfo:
      type: object
      properties:
        code:
          type: integer
          description: Код ошибки
          example: 1001
        message:
          type: string
          description: Описание ошибки
          example: "Client not found"
        details:
          type: string
          description: Детальная информация об ошибке
          example: "Client with ID Api-clientId-01 does not exist"
`

	// Invalid provider specification (with breaking changes)
	providerSpecInvalid := `asyncapi: 3.0.0
info:
  title: Wallet Balance Service
  version: 1.0.0
  description: |
    Сервис для получения баланса электронного кошелька клиента с ошибками в структуре.

servers:
  production:
    host: localhost:8080
    protocol: https
    description: Production server для wallet balance service

channels:
  walletBalanceRequest:
    address: "/api/client/{clientId}/wallet/balance"
    description: |
      Канал для получения запросов на баланс кошелька клиента.
    messages:
      getBalanceRequest:
        $ref: '#/components/messages/getBalanceRequest'
  
  walletBalanceResponse:
    address: "/api/client/{clientId}/wallet/balance"
    description: |
      Канал для отправки ответов с информацией о балансе кошелька.
    messages:
      getBalanceResponse:
        $ref: '#/components/messages/getBalanceResponse'

operations:
  receiveBalanceRequest:
    action: receive
    channel:
      $ref: "#/channels/walletBalanceRequest"
    
  sendBalanceResponse:
    action: send
    channel:
      $ref: "#/channels/walletBalanceResponse"

components:
  messages:
    getBalanceRequest:
      name: GetBalanceRequest
      title: Запрос на получение баланса кошелька
      contentType: application/json
      headers:
        $ref: '#/components/schemas/requestHeaders'
      payload:
        $ref: '#/components/schemas/balanceRequestPayload'
    
    getBalanceResponse:
      name: GetBalanceResponse  
      title: Ответ с балансом кошелька
      contentType: application/json
      headers:
        $ref: '#/components/schemas/responseHeaders'
      payload:
        $ref: '#/components/schemas/balanceResponsePayload'

  schemas:
    requestHeaders:
      type: object
      properties:
        Api-clientId:
          type: string
          description: Идентификатор клиента из URL path
          example: "Api-clientId-01"
        RequestID:
          type: string
          description: Уникальный идентификатор запроса
          example: "request-01"
        Content-Type:
          type: string
          enum: ["application/json"]
          example: "application/json"
        # Breaking change: added new required field
        NewRequiredField:
          type: string
          description: Новое обязательное поле
      required:
        - Api-clientId
        - Content-Type
        - NewRequiredField  # Breaking change

    responseHeaders:
      type: object
      properties:
        Content-Type:
          type: string
          enum: ["application/json"]
          example: "application/json"
        RequestID:
          type: string
          description: Идентификатор запроса для корреляции
          example: "request-01"
      required:
        - Content-Type

    balanceRequestPayload:
      type: object
      required:
        - requestType
        - dateFrom
        - dateTo
        - operTime
        - newRequiredField  # Breaking change: new required field
      properties:
        requestType:
          type: string
          description: Тип запроса для получения истории операций
          enum:
            - operationHistory
            - newOperation  # Breaking change: new enum value that consumer doesn't expect
          example: "operationHistory"
        dateFrom:
          type: string  # Breaking change: changed from number to string
          format: date-time
          description: |
            Дата начала периода в ISO format
          example: "2025-01-15T10:30:00Z"
        dateTo:
          type: string  # Breaking change: changed from number to string
          format: date-time
          description: |
            Дата окончания периода в ISO format
          example: "2025-01-15T11:30:00Z"
        operTime:
          type: integer  # Breaking change: changed from number to integer
          format: int64
          description: |
            Время операции в миллисекундах
          example: 1745299737198
        newRequiredField:
          type: string
          description: Новое обязательное поле

    balanceResponsePayload:
      type: object
      required:
        - status
        - actualTimestamp
        - data
      properties:
        status:
          type: string
          description: Статус выполнения запроса
          enum:
            - success
            - error
            - pending  # Breaking change: new enum value
          example: "success"
        actualTimestamp:
          type: string  # Breaking change: changed from integer to string
          format: date-time
          description: |
            Актуальное время ответа в ISO format
          example: "2025-01-15T10:30:00Z"
        data:
          $ref: '#/components/schemas/walletBalanceData'
        error:
          $ref: '#/components/schemas/errorInfo'

    walletBalanceData:
      type: object
      required:
        - createTime
        - operationId
        - balance
      properties:
        createTime:
          type: string  # Breaking change: changed from integer to string
          format: date-time
          description: |
            Время создания записи о балансе в ISO format
          example: "2025-01-15T10:30:00Z"
        operationId:
          type: integer  # Breaking change: changed from string to integer
          description: Уникальный идентификатор операции
          example: 1
        balance:
          $ref: '#/components/schemas/balanceInfo'

    balanceInfo:
      type: object
      required:
        - value
        - currency
      properties:
        value:
          type: integer  # Breaking change: changed from number to integer
          description: Сумма баланса в копейках
          example: 99990
        currency:
          type: string
          description: Код валюты баланса
          enum:
            - rub
            - usd
            - eur
            - btc  # Breaking change: new enum value
          example: "rub"

    errorInfo:
      type: object
      properties:
        code:
          type: string  # Breaking change: changed from integer to string
          description: Код ошибки
          example: "E1001"
        message:
          type: string
          description: Описание ошибки
          example: "Client not found"
        details:
          type: string
          description: Детальная информация об ошибке
          example: "Client with ID Api-clientId-01 does not exist"
`

	// Write test data files
	if err := os.WriteFile("testdata/consumer_spec.yml", []byte(consumerSpec), 0644); err != nil {
		t.Fatalf("Failed to write consumer spec: %v", err)
	}

	if err := os.WriteFile("testdata/provider_spec_valid.yml", []byte(providerSpecValid), 0644); err != nil {
		t.Fatalf("Failed to write valid provider spec: %v", err)
	}

	if err := os.WriteFile("testdata/provider_spec_invalid.yml", []byte(providerSpecInvalid), 0644); err != nil {
		t.Fatalf("Failed to write invalid provider spec: %v", err)
	}
}

func cleanupTestData() {
	os.RemoveAll("testdata")
}

// Benchmark tests
func BenchmarkContractCompatibilityChecker_CheckCompatibility(b *testing.B) {
	setupTestData(&testing.T{})
	defer cleanupTestData()

	checker := NewContractCompatibilityChecker(
		"testdata/consumer_spec.yml",
		"testdata/provider_spec_valid.yml",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.CheckCompatibility()
		if err != nil {
			b.Fatalf("CheckCompatibility() error = %v", err)
		}
	}
}

func BenchmarkContractCompatibilityChecker_parseYAMLToJSON(b *testing.B) {
	yamlContent := []byte(`asyncapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
channels:
  test:
    address: "/api/test"
    messages:
      testMessage:
        $ref: "#/components/messages/testMessage"
components:
  messages:
    testMessage:
      name: TestMessage
      contentType: application/json
      payload:
        $ref: "#/components/schemas/testSchema"
  schemas:
    testSchema:
      type: object
      properties:
        id:
          type: string
        value:
          type: number`)

	checker := &ContractCompatibilityChecker{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.parseYAMLToJSON(yamlContent)
		if err != nil {
			b.Fatalf("parseYAMLToJSON() error = %v", err)
		}
	}
}