# mq-rest-sync-adapter

Пример рефакторинга по кукбуку.

В серии статей я опишу, почему я прошу разработчиков читать книги  
В частности хорошие книги серии Дядюшки Боба (Роберта Мартина)

# Назначение сервиса:
Сервис является синхронным перекладчиком запросов систем-потребителей в системы-поставщики.  
Запросы поступают из систем-потребителей по очередям и обрабатываются сервисом-поставщиком согласно обобщенному бизнес-процессу:  
Запросы из очереди направляются синхронно по REST'у в соответствующую систему, полученный ответ упаковывается в ответный конверт и отправляется в очередь ответа. Подробнее в описании каждого процесса.

## Обработка запроса из Платежной Системы (PipelineServicePS.kt)

| Получает конверт с запросом  
| Валидирует входящий запрос  
| Отправляет квитанцию о результате валидации и принятии запроса  
| Отправляет запрос на получение истории операций в Систему А (REST)  
| Отправляет полученный ответ в очередь  
| Отправляет квитанцию об успехе либо ошибке обработки запроса  

![pipeline](https://www.plantuml.com/plantuml/png/bL9BIiDG59rd5NUeMz199sv0tO2WCpzYuxnfg1yiHbo0uWvUGXvzfirp2-VkoFCkeH2m0w65U-_vdEIGoJG_kyllhoup6x83V29YhQK-mi7hND3nqAFXCkEgXM15w4TdY5eGsnnIzSxJ6W4j0ccbAT7eWdYhbJuJ-Xu9a-X3vxOwD8oiXZWkWO13hm-SNWLHIF9OQKdQCc7yU_flGrsY8WX_vXWehaZYRPGAjoDtn1BqS3JQDk5v22anQ45j9CsS8pcinUBj1Hl1rq3os6WfU9EzuwEJq_sTkUEKNb-IJoIp6KfC6q6heDNbluf3BRGkx34Ny69rrOY4zCd63jvuvbPRSFT7QTFyvZQRN5oVyrFv1G00)


