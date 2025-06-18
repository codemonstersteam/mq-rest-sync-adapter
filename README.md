# mq-rest-sync-adapter


## Назначение сервиса:
Сервис является синхронным перекладчиком запросов систем-потребителей в системы-поставщики.  
Запросы поступают из систем-потребителей по очередям и обрабатываются сервисом-поставщиком согласно обобщенному бизнес-процессу:  
Запросы из очереди направляются синхронно по REST'у в соответствующую систему, полученный ответ упаковывается в ответный конверт и отправляется в очередь ответа. Подробнее в описании каждого процесса.

## Обработка запроса из Платежной Системы (PipelineServicePS.kt)

| Получает конверт с запросом  
| Валидирует входящий запрос  
| Отправляет квитанцию о результате валидации и принятии запроса  
| Отправляет запрос на получение баланса кошелька в Систему А (REST)  
| Отправляет полученный ответ в очередь  
| Отправляет квитанцию об успехе либо ошибке обработки запроса  

```plantuml

start
:Получает конверт с запросом;
:Валидирует входящий запрос;
:Отправляет квитанцию о результате валидации и принятии запроса;
:Отправляет запрос на получение истории операций в Систему А (REST);
:Отправляет полученный ответ в очередь;
:Отправляет квитанцию об успехе либо ошибке обработки запроса;
stop

```

## Полезные ссылки

- [How to build local](HOW-TO-BUILD-LOCAL.md)
- [How to run component tests](component-tests/README.md)


# Развитие проекта
Начало истории в [01-dev-origin](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/tree/01-dev-origin?ref_type=heads)

Этот [выразительный пример плохого качества](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/tree/01-dev-origin?ref_type=heads) кодописи и применения [кукбука](https://git.codemonsters.team/guides/ddd-code-toolkit) мотивировал меня писать статьи про рефакторинг.
- [Как написать чистый код и сделать жизнь проще](https://codemonsters.team/effective-developer/code/clean-code/how-to/)  
- [Как не стоит писать код: разбираем ошибки](https://codemonsters.team/effective-developer/code/clean-code/how-not-to-write-code/)  
- [Гайд по чистому коду: начало улучшений. Какие тесты добавить в проект в первую очередь](https://codemonsters.team/effective-developer/code/clean-code/how-to-integration-test/)  

Далее я описывал рефакторинг этого проекта.
Я осознал в какой-то момент, что хочу написать электрокнигу для введения в эффективную разработку
Статья получилась большой и сложной для восприятия новичка.
Драфт большой статьи я выложу. На свой сайт в блог

Ты читаешь исходник сервиса с переписанным куском бизнес-логики.
Этот улучшенный кусок [встроен](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/blob/feature/02-vertical-slice-refactoring/src/main/kotlin/team/codemonsters/refactoringexample/service/PipelineServicePS.kt?ref_type=heads#L53) по вертикальному срезу с фича тоглом

## Бранчи и история 
- [0 01-dev-origin: исходное состояние проекта](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/tree/01-dev-origin?ref_type=heads)  
- [01 01-readme: пишем основу - README.md](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/tree/feature/01-readme?ref_type=heads)  
- [01 01-integration-tests : Пишем компонентный тест на бизнес-процесс](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/tree/feature/01-integration-tests?ref_type=heads)  
- [02 02-vertical-slice : Описываем новый процесс в вертикальном срезе и пузырьке(Bubble context)](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/tree/feature/02-vertical-slice?ref_type=heads)  
  - [ССылка на бизнес-логику в вертикальном срезе](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/tree/feature/02-vertical-slice/src/main/kotlin/team/codemonsters/refactoringexample/walletBalance?ref_type=heads)  
  - [Ссылка на легаси точку входа в бизнес-процесс и фича тогл](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/blob/feature/02-vertical-slice/src/main/kotlin/team/codemonsters/refactoringexample/service/PipelineServicePS.kt?ref_type=heads)  
- [02-vertical-slice-refactoring : Улучшенный код процесса с сохранением легаси](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/tree/feature/02-vertical-slice-refactoring?ref_type=heads)  
- [main : кусок кода, который стремится к эталону (вмерджил 04.05)](https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/tree/main?ref_type=heads)   


## Пример рефакторинга по кукбуку

В серии статей я опишу, почему я прошу разработчиков читать книги  
В частности хорошие книги серии Дядюшки Боба (Роберта Мартина)

