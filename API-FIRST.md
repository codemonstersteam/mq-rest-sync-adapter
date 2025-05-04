# API First

## Async API
https://www.asyncapi.com/docs/tools/cli/usage

```bash
asyncapi validate api-specification/asyncapi.yaml
```

## Docker
https://hub.docker.com/r/asyncapi/

https://hub.docker.com/r/asyncapi/generator

```bash
docker run --rm -it \
--user=root \
-v api-specification/asyncapi.yaml:/app/asyncapi.yml \
-v build:/app/output \
asyncapi/cli --help
```

```bash
docker run --rm -it \
--entrypoint='' \
--user=root \
-v ./api-specification/asyncapi.yml:/app/asyncapi.yml \
-v ./build:/app/output \
asyncapi/cli:3.1.1 \
asyncapi validate /app/asyncapi.yml
```

22.04.2025
Разобрался с async api 3.0.0
Пусть мне мой ИИ Друг поможет написать остаток спеки на исходящий запрос

```http
Напиши async api specification версии 3.0.0
на исходящий запрос из сервиса
адрес канала url: http://localhost:40303/api/client/Api-clientId-01/wallet/balance
http заголовок:
{Api-clientId=Api-clientId-01, RequestID=request-01}
сервис в который уходит запрос защищен basicAuth user=admin password=password
тело запроса исходящего из сервиса:
{"requestType":"operationHistory","dateFrom":1745299787.198193738,"dateTo":1745299792.198546764,"operTime":1745299737.198562113}
```

Результат:
```yaml
asyncapi: 3.0.0
info:
  title: Wallet Balance API
  version: 1.0.0
  description: API для запроса истории операций кошелька

servers:
  production:
    url: http://localhost:40303
    protocol: http
    security:
      - basicAuth: []
    variables:
      clientId:
        default: Api-clientId-01

channels:
  '/api/client/{clientId}/wallet/balance':
    address: 
      path: /api/client/{clientId}/wallet/balance
    messages:
      walletBalanceRequest:
        name: WalletBalanceRequest
        title: Запрос истории операций
        summary: Запрос истории операций кошелька
        contentType: application/json
        headers:
          type: object
          properties:
            Api-clientId:
              type: string
              example: Api-clientId-01
            RequestID:
              type: string
              example: request-01
        payload:
          type: object
          properties:
            requestType:
              type: string
              example: "operationHistory"
            dateFrom:
              type: number
              format: double
              example: 1745299787.198193738
            dateTo:
              type: number
              format: double
              example: 1745299792.198546764
            operTime:
              type: number
              format: double
              example: 1745299737.198562113
          required:
            - requestType
            - dateFrom
            - dateTo
            - operTime

operations:
  sendWalletBalanceRequest:
    action: send
    channel:
      $ref: '#/channels//api/client/{clientId}/wallet/balance'
    messages:
      - $ref: '#/channels//api/client/{clientId}/wallet/balance/messages/walletBalanceRequest'

components:
  securitySchemes:
    basicAuth:
      type: http
      scheme: basic
      description: Basic authentication
```