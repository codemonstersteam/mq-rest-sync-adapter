# Обзор проекта
Тестирую полный цикл работы с ассистентами.
В данном доке я описываю процесс создания проекта для компонентного тестирования.

## Запуск тестов

````bash
export FT_ENABLED=true # true - включить новый поток, 
export FT_ENABLED=false # false - работает старый поток
./scripts/run-test-sut-v2.sh
````
Запуск тестов можно улучшить:  
Добавить в геркин скрипты для установки значений тоглов
Использовать буду удаленное хранилище для тоглов (Unleash или подобный инструмент)  
Выберем с машиной подходящий:)  


## LOG Работы с ассистентами

07.05.2025 20:07
за два часа с нуля создал проект для компонентного тестирования.
с проверкой базовых зависимостей 
ibm mq, wiremock
с использованием https://claude.ai/
версия Claude 3.7 Sonnet
пришлось шлифовать, чтобы получился нормальный проект.
Но в целом работа выполнена быстро.
с решением проблем помогал cursor

завтра подниму все скриптом и буду проверять.
Важно, чтобы проект был написан на простом Java, без всяких там Spring.

08.05.2025 
за 45 минут исправил с курсором и доработал тест на wiremock.
за 2 часа доработал тест на ibm mq.
Помогли два ассистента курсор и deep seek, про cloude я забыл в моменте.

Все готово для создания полноценного компонентного теста приложения

09.05 полный запуск проверки среды тестирования для компонента готово
Доработал документацию и gitlab ci
Лучше разобрался с VS Code шорткатами клавиатуры
Нравится все больше. не смог вернуться в Idea, попытка была. Начал искать плагины,
кривой плагин для deep seek вывел меня из себя.

# Log проекта

Эксперимент поставил следующий.
Я с нуля с использованием Claude 3.7 Sonnet и cursor создал проект для компонентного тестирования.
Вот запрос:
```text
I am writing a guide for creating simple component tests at the merge request stage
The guide needs to be written using java, cucumber, gradle, docker
I need to create a separate project in git that will contain everything needed to test the service in isolation using cucumber gherkin
use gradle as a project builder
in docker, raise the assembled service from the main branch of git with the merged changes, 
here is the source code of the service
https://git.codemonsters.team/guides/mq-rest-sync-adapter/
Next, you need to
raise all the dependencies necessary for this service in docker
submit data to the input of the transport dependency,
in our case, send a request to the IBM queue,
pay attention to the specification service
https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/blob/main/api-specification/asyncapi.yml?ref_type=heads
get expected response from queue
this is testing service in isolation with stubs
```
следующий промпт:
```text
 we can build the required version of the service merge into the main branch commit from the merge request
then use the required service docker in docker compose
with the correct configs for stubs, which we will raise in docker
and then write tests on gherkin and run them
```
следующий промпт:

```text
you are missing that the mq-rest-sync-adapter service has a dependency
according to the specification
https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/blob/main/api-specification/asyncapi.yml?ref_type=heads
restServer line 31
channel getBalanceRequest line 78
the service requests data from the rest service from channel getBalanceRequest
and send response to response channel sendRestResponse to OUT.QUEUE.PS
i.e. we have two dependencies, 
IBM MQ and rest service
the stub of which we need to raise using wireMock 
and we can use wireMock stubs from test source of service https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/tree/main/src/test/resources/wiremock/rest-api-gateway/mappings?ref_type=heads
```
и еще корректировка:
```text
change package from  com.codemonsters to team.codemonsters as in source project
https://git.codemonsters.team/guides/mq-rest-sync-adapter

take the content of wiremock/mappings/getBalanceRequest.json
from the source of service https://git.codemonsters.team/guides/mq-rest-sync-adapter/-/blob/main/src/test/resources/wiremock/rest-api-gateway/mappings/wallet-balance.json?ref_type=heads
```
клод не смог забрать контент из ссылки, поэтому пришлось вручную вставить в файл.
```
take take the content of wiremock/mappings/getBalanceRequest.json
from this content

{
  "mappings": [
    {
      "request": {
        "method": "POST",
        "url": "/api/client/Api-clientId-01/wallet/balance"
      },
      "response": {
        "status": 200,
        "headers": {
          "Content-Type": "application/json"
        },
        "jsonBody": {
          "status": "success",
          "actualTimestamp": 1607449018829,
          "data": {
            "createTime": 1600443296227,
            "operationId": "01",
            "balance": {
              "value": 999.9,
              "currency": "rub"
            }
          }
        }
      }
    }
  ]
}
```
Далее я перешел в дипсик:
Чтобы настроить ibm mq в docker и разрешить проблемы тестирования работы интеграции с ibm mq:
```
Смотри Вот компоуз

version: '3.8'

services:
  # IBM MQ Service
  ibmmq:
    image: ibmcom/mq:9.2.4.0-r1-amd64
    extra_hosts:
      - "locahost:0.0.0.0"
    container_name: ibmmq
    ports:
      - "1414:1414"
      - "9443:9443"
    environment:
      - LICENSE=accept
      - MQ_QMGR_NAME=QM1
      - MQ_DEV=true
    networks:
      - test-network
    volumes:
      - ./scripts/mqsc/20-config.mqsc:/etc/mqm/20-config.mqsc
    healthcheck:
      test: ["CMD-SHELL", "chkmqhealthy || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5

  # WireMock REST API Stub
  wiremock:
    image: wiremock/wiremock:3.13.0
    container_name: wiremock
    ports:
      - "8080:8080"
    volumes:
      - ./wiremock:/home/wiremock
    command: 
      - --global-response-templating
      - --verbose
    networks:
      - test-network
    healthcheck:
      test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/__admin/mappings || exit 1"]
      interval: 5s
      timeout: 3s
      retries: 5

  # MQ-REST-Sync-Adapter Service
  

networks:
  test-network:
    driver: bridge

При подключении с параметрами 

MQQueueConnectionFactory
 
            factory = new MQQueueConnectionFactory();
            factory.setHostName("localhost");
            factory.setPort("1414");
            factory.setQueueManager("QM1");
            factory.setChannel("DEV.ADMIN.SVRCONN");
            connection = factory.createConnection("admin","passw0rd");

выдает ошибку
com.ibm.msg.client.jms.DetailedJMSException: JMSFMQ6312: An exception occurred in the Java(tm) MQI.
```
Результат:
```yml
version: '3.8'

services:
  ibmmq:
    image: ibmcom/mq:9.2.4.0-r1-amd64
    hostname: ibmmq  # Важно для корректной работы MQ!
    container_name: ibmmq
    ports:
      - "1414:1414"
      - "9443:9443"
    environment:
      - LICENSE=accept
      - MQ_QMGR_NAME=QM1
      - MQ_DEV=true
      - MQ_ADMIN_PASSWORD=passw0rd
      - MQ_APP_PASSWORD=passw0rd
      - MQ_CHANNEL=DEV.APP.SVRCONN  # Канал по умолчанию
    networks:
      - test-network
    volumes:
      - ./scripts/mqsc/20-config.mqsc:/etc/mqm/20-config.mqsc
    healthcheck:
      test: ["CMD-SHELL", "echo 'DISPLAY CHSTATUS(DEV.APP.SVRCONN)' | runmqsc QM1 | grep -i 'RUNNING' || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5

  wiremock:
    # ... (остается без изменений)

networks:
  test-network:
    driver: bridge

```  
И код подключения к ibm mq:

```java
MQQueueConnectionFactory factory = new MQQueueConnectionFactory();
factory.setHostName("ibmmq");  // Используем имя сервиса из docker-compose
factory.setPort(1414);
factory.setQueueManager("QM1");
factory.setChannel("DEV.APP.SVRCONN");  // Должен совпадать с MQ_CHANNEL в компоуз
factory.setTransportType(WMQConstants.WMQ_CM_CLIENT);  // Обязательно для Docker

// Используем app пользователя, а не admin
Connection connection = factory.createConnection("app", "passw0rd");
```
Далаа дипсик помог мне пофикисить проблемы с правами доступа к очередям.
```
смотри на 20-config.mqsc

DEFINE CHANNEL(DEV.APP.SVRCONN) CHLTYPE(SVRCONN) TRPTYPE(TCP)
SET CHLAUTH(DEV.APP.SVRCONN) TYPE(BLOCKUSER) USERLIST('nobody')
ALTER AUTHINFO(SYSTEM.DEFAULT.AUTHINFO.IDPWOS) AUTHTYPE(IDPWOS) ADOPTCTX(YES)
REFRESH SECURITY TYPE(CONNAUTH)

DEFINE QLOCAL(MY.TEST) REPLACE
DEFINE QLOCAL(IN.QUEUE.PS) REPLACE
DEFINE QLOCAL(OUT.QUEUE.PS) REPLACE
DEFINE QLOCAL(RECEIPT.QUEUE.PS) REPLACE
DEFINE QLOCAL(IN.QUEUE.MID) REPLACE

теперь выдает ошибку
com.ibm.msg.client.jms.DetailedJMSSecurityException: JMSWMQ2008: Failed to open MQ queue 'IN.QUEUE.PS'.
```
Результат рабочий конфиг очередей:
```text
DEFINE CHANNEL(DEV.APP.SVRCONN) CHLTYPE(SVRCONN) TRPTYPE(TCP)
SET CHLAUTH(DEV.APP.SVRCONN) TYPE(BLOCKUSER) USERLIST('nobody')
ALTER AUTHINFO(SYSTEM.DEFAULT.AUTHINFO.IDPWOS) AUTHTYPE(IDPWOS) ADOPTCTX(YES)
REFRESH SECURITY TYPE(CONNAUTH)

* Определение очередей
DEFINE QLOCAL(MY.TEST) REPLACE
DEFINE QLOCAL(IN.QUEUE.PS) REPLACE
DEFINE QLOCAL(OUT.QUEUE.PS) REPLACE
DEFINE QLOCAL(RECEIPT.QUEUE.PS) REPLACE
DEFINE QLOCAL(IN.QUEUE.MID) REPLACE

* Настройка прав доступа
SET AUTHREC PROFILE('IN.QUEUE.PS') OBJTYPE(QUEUE) PRINCIPAL('app') AUTHADD(BROWSE,GET,PUT,INQ)
SET AUTHREC PROFILE('OUT.QUEUE.PS') OBJTYPE(QUEUE) PRINCIPAL('app') AUTHADD(BROWSE,GET,PUT,INQ)
SET AUTHREC PROFILE('SYSTEM.DEFAULT.MODEL.QUEUE') OBJTYPE(QUEUE) PRINCIPAL('app') AUTHADD(BROWSE,GET,PUT,INQ)

* Разрешить доступ к менеджеру очередей
SET AUTHREC PROFILE('QM1') OBJTYPE(QMGR) PRINCIPAL('app') AUTHADD(CONNECT,INQ)

REFRESH SECURITY
```

### И объяснение каждой строки:
тут будет текст от дипсика


# Component Testing Guide for MQ-REST-Sync-Adapter at Merge Request Stage

This guide outlines how to create isolated component tests for the MQ-REST-Sync-Adapter service using Docker Compose, WireMock, and Gherkin/Cucumber with Gradle as the build system.

## Overview

The MQ-REST-Sync-Adapter is a service that bridges between IBM MQ queues and a REST API. Based on the AsyncAPI specification, the service:

1. Listens to messages on an input MQ queue
2. Makes a REST API call to a backend service to retrieve data (e.g., getBalanceRequest)
3. Returns the response via an output MQ queue (OUT.QUEUE.PS)

Our testing approach will:

1. Create a separate Git project specifically for component testing
2. Use Docker Compose to orchestrate three components:
   - IBM MQ (message broker)
   - WireMock (REST API stub)
   - MQ-REST-Sync-Adapter service (built from the main branch with merged changes)
3. Implement Cucumber Gherkin scenarios to test the service behavior
4. Reuse existing WireMock stubs from the service's test resources

## Project Structure

```
mq-rest-sync-adapter-component-tests/
├── build.gradle
├── settings.gradle
├── gradle/
│   └── wrapper/
├── docker-compose.yml
├── scripts/
│   ├── build-service-image.sh
│   └── run-tests.sh
├── src/
│   ├── main/
│   │   ├── java/
│   │   │   └── team/
│   │   │       └── codemonsters/
│   │   │           └── mqrestsyncadapter/
│   │   │               └── test/
│   │   │                   ├── config/
│   │   │                   ├── client/
│   │   │                   └── utils/
│   │   └── resources/
│   │       └── wiremock/
│   │           ├── mappings/
│   │           │   └── getBalanceRequest.json
│   │           └── __files/
│   │               └── responses/
│   └── test/
│       ├── java/
│       │   └── team/
│       │       └── codemonsters/
│       │           └── mqrestsyncadapter/
│       │               └── test/
│       │                   ├── stepdefs/
│       │                   └── CucumberTestRunner.java
│       └── resources/
│           ├── features/
│           │   └── message_processing.feature
│           └── cucumber.properties
├── wiremock/
│   ├── mappings/
│   │   └── getBalanceRequest.json
│   └── __files/
│       └── responses/
└── README.md
```

## Setup Instructions

### 1. Create a New Gradle Project

First, create a new Gradle project specifically for component testing:

```bash
mkdir mq-rest-sync-adapter-component-tests
cd mq-rest-sync-adapter-component-tests
gradle init --type java-application --dsl groovy --package team.codemonsters.mqrestsyncadapter.test
```

### 2. Configure Gradle Dependencies

Edit the `build.gradle` file to include necessary dependencies:

```groovy
plugins {
    id 'java'
}

group = 'team.codemonsters.mqrestsyncadapter.test'
version = '1.0.0-SNAPSHOT'
sourceCompatibility = '21'

repositories {
    mavenCentral()
}

configurations {
    cucumberRuntime {
        extendsFrom testImplementation
    }
}

dependencies {
    // IBM MQ Client 
    implementation 'com.ibm.mq:com.ibm.mq.allclient:9.3.3.0'
    
    // Cucumber for BDD testing
    testImplementation 'io.cucumber:cucumber-java:7.14.0'
    testImplementation 'io.cucumber:cucumber-junit:7.14.0'
    
    // REST client for testing REST endpoints and WireMock admin API
    implementation 'org.apache.httpcomponents:httpclient:4.5.14'
    
    // JSON Processing
    implementation 'org.json:json:20231013'
    
    // JUnit and assertion libraries
    testImplementation 'org.junit.jupiter:junit-jupiter-api:5.9.3'
    testRuntimeOnly 'org.junit.jupiter:junit-jupiter-engine:5.9.3'
    testImplementation 'org.assertj:assertj-core:3.24.2'
    
    // Logging
    implementation 'org.slf4j:slf4j-api:2.0.7'
    implementation 'ch.qos.logback:logback-classic:1.4.8'
}

tasks.register('cucumber', JavaExec) {
    dependsOn assemble, testClasses
    classpath = configurations.testRuntimeClasspath + sourceSets.main.output + sourceSets.test.output
    mainClass = "io.cucumber.core.cli.Main"
    args = [
        '--plugin', 'pretty',
        '--plugin', 'html:build/reports/cucumber.html',
        '--plugin', 'json:build/reports/cucumber.json',
        '--glue', 'team.codemonsters.mqrestsyncadapter.test.stepdefs',
        'src/test/resources/features'
    ]
}


test {
    useJUnitPlatform()
    finalizedBy cucumber
}
```

### 3. Create Docker Compose Configuration

Create a `docker-compose.yml` file for the complete test environment:

```yaml
version: '3.8'

services:
  # IBM MQ Service
  ibmmq:
    image: ibmcom/mq:9.2.4.0-r1-amd64
    hostname: ibmmq  # Важно для корректной работы MQ!
    container_name: ibmmq
    ports:
      - "1414:1414"
      - "9443:9443"
    environment:
      - LICENSE=accept
      - MQ_QMGR_NAME=QM1
      - MQ_DEV=true
      - MQ_ADMIN_PASSWORD=passw0rd
      - MQ_APP_PASSWORD=passw0rd
      - MQ_CHANNEL=DEV.APP.SVRCONN  # Канал по умолчанию
    networks:
      - test-network
    volumes:
      - ./scripts/mqsc/20-config.mqsc:/etc/mqm/20-config.mqsc
    healthcheck:
      test: ["CMD-SHELL", "chkmqhealthy || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5

  # WireMock REST API Stub
  wiremock:
    image: wiremock/wiremock:3.13.0
    container_name: wiremock
    ports:
      - "8080:8080"
    volumes:
      - ./wiremock:/home/wiremock
    command: 
      - --global-response-templating
      - --verbose
    networks:
      - test-network
    healthcheck:
      test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/__admin/mappings || exit 1"]
      interval: 5s
      timeout: 3s
      retries: 5

  # MQ-REST-Sync-Adapter Service
  

networks:
  test-network:
    driver: bridge

```

### 4. Create MQ Configuration

Create a directory for MQ configuration scripts:

```bash
mkdir -p scripts/mqsc
```

Create a file `scripts/mqsc/config.mqsc` with the following content:

```text
DEFINE CHANNEL(DEV.APP.SVRCONN) CHLTYPE(SVRCONN) TRPTYPE(TCP)
SET CHLAUTH(DEV.APP.SVRCONN) TYPE(BLOCKUSER) USERLIST('nobody')
ALTER AUTHINFO(SYSTEM.DEFAULT.AUTHINFO.IDPWOS) AUTHTYPE(IDPWOS) ADOPTCTX(YES)
REFRESH SECURITY TYPE(CONNAUTH)

* Определение очередей
DEFINE QLOCAL(MY.TEST) REPLACE
DEFINE QLOCAL(IN.QUEUE.PS) REPLACE
DEFINE QLOCAL(OUT.QUEUE.PS) REPLACE
DEFINE QLOCAL(RECEIPT.QUEUE.PS) REPLACE
DEFINE QLOCAL(IN.QUEUE.MID) REPLACE

* Настройка прав доступа
SET AUTHREC PROFILE('IN.QUEUE.PS') OBJTYPE(QUEUE) PRINCIPAL('app') AUTHADD(BROWSE,GET,PUT,INQ)
SET AUTHREC PROFILE('OUT.QUEUE.PS') OBJTYPE(QUEUE) PRINCIPAL('app') AUTHADD(BROWSE,GET,PUT,INQ)
SET AUTHREC PROFILE('SYSTEM.DEFAULT.MODEL.QUEUE') OBJTYPE(QUEUE) PRINCIPAL('app') AUTHADD(BROWSE,GET,PUT,INQ)

* Разрешить доступ к менеджеру очередей
SET AUTHREC PROFILE('QM1') OBJTYPE(QMGR) PRINCIPAL('app') AUTHADD(CONNECT,INQ)

REFRESH SECURITY
```

### 5. Set up WireMock Stubs

Create the directories for WireMock stubs:

```bash
mkdir -p wiremock/mappings
mkdir -p wiremock/__files/responses
```

Create a WireMock stub for the wallet balance endpoint in `wiremock/mappings/wallet-balance.json`:

```json
{
  "mappings": [
    {
      "request": {
        "method": "POST",
        "url": "/api/client/Api-clientId-01/wallet/balance"
      },
      "response": {
        "status": 200,
        "headers": {
          "Content-Type": "application/json"
        },
        "jsonBody": {
          "status": "success",
          "actualTimestamp": 1607449018829,
          "data": {
            "createTime": 1600443296227,
            "operationId": "01",
            "balance": {
              "value": 999.9,
              "currency": "rub"
            }
          }
        }
      }
    }
  ]
}
```


### 7. Create a WireMock Client Helper

Create a class to interact with the WireMock REST API:

```java
package team.codemonsters.mqrestsyncadapter.test.client;

import org.apache.http.HttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;

public class WireMockClient {
    private static final Logger LOG = LoggerFactory.getLogger(WireMockClient.class);
    
    private final String baseUrl;
    
    public WireMockClient(String host, int port) {
        this.baseUrl = "http://" + host + ":" + port;
    }
    
    public boolean isRunning() {
        try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
            HttpGet request = new HttpGet(baseUrl + "/__admin/mappings");
            HttpResponse response = httpClient.execute(request);
            return response.getStatusLine().getStatusCode() == 200;
        } catch (IOException e) {
            LOG.error("Error checking WireMock status", e);
            return false;
        }
    }
    
    public String getRequests() {
        try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
            HttpGet request = new HttpGet(baseUrl + "/__admin/requests");
            HttpResponse response = httpClient.execute(request);
            if (response.getStatusLine().getStatusCode() == 200) {
                return EntityUtils.toString(response.getEntity());
            } else {
                LOG.error("Error getting WireMock requests: {}", response.getStatusLine());
                return null;
            }
        } catch (IOException e) {
            LOG.error("Error getting WireMock requests", e);
            return null;
        }
    }
    
    public void resetRequests() {
        try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
            HttpGet request = new HttpGet(baseUrl + "/__admin/requests/reset");
            httpClient.execute(request);
        } catch (IOException e) {
            LOG.error("Error resetting WireMock requests", e);
        }
    }
}
```

### 8. Create A Test Configuration Helper

Create a TestConfig class to hold environment-specific settings:

```java
package team.codemonsters.mqrestsyncadapter.test.utils;

public class TestConfig {
    // MQ Configuration
    public static final String MQ_HOST = getEnv("MQ_HOST", "localhost");
    public static final int MQ_PORT = Integer.parseInt(getEnv("MQ_PORT", "1414"));
    public static final String MQ_QMGR_NAME = getEnv("MQ_QMGR_NAME", "QM1");
    public static final String MQ_CHANNEL = getEnv("MQ_CHANNEL", "DEV.APP.SVRCONN");
    public static final String MQ_USER = getEnv("MQ_USER", "app");
    public static final String MQ_PASSWORD = getEnv("MQ_PASSWORD", "passw0rd");
    
    // WireMock Configuration
    public static final String WIREMOCK_HOST = getEnv("WIREMOCK_HOST", "localhost");
    public static final int WIREMOCK_PORT = Integer.parseInt(getEnv("WIREMOCK_PORT", "8080"));
    
    private static String getEnv(String key, String defaultValue) {
        String value = System.getenv(key);
        return value != null ? value : defaultValue;
    }
} 
```

### 9. Define Cucumber Features

Create Gherkin feature files to describe the test scenarios:

```gherkin

```

### 10. Implement Step Definitions

Create step definitions to implement the Cucumber scenarios:

```java

```

### 11. Create the Cucumber Test Runner

```java
package team.codemonsters.mqrestsyncadapter.test;

import io.cucumber.junit.Cucumber;
import io.cucumber.junit.CucumberOptions;
import org.junit.runner.RunWith;

@RunWith(Cucumber.class)
@CucumberOptions(
    features = "src/test/resources/features",
    glue = "team.codemonsters.mqrestsyncadapter.test.stepdefs",
    plugin = {"pretty", "html:build/reports/cucumber/report.html"}
)
public class CucumberTestRunner {
}
```

## Docker Image Build Script

Create a script to build the service image with merged changes:

```bash
#!/bin/bash
# scripts/build-service-image.sh

# Clone the main repository if it doesn't exist
if [ ! -d "mq-rest-sync-adapter" ]; then
  git clone https://git.codemonsters.team/guides/mq-rest-sync-adapter.git
fi

# Navigate to the repository
cd mq-rest-sync-adapter

# Switch to main branch
git checkout main

# Pull latest changes
git pull

# If there's a merge request ID provided, fetch and merge it
if [ ! -z "$MR_ID" ]; then
  echo "Fetching and merging changes from merge request #$MR_ID"
  git fetch origin merge-requests/$MR_ID/head:mr-$MR_ID
  git checkout mr-$MR_ID
fi

# Build the Docker image
docker build -t mq-rest-sync-adapter:latest .

# Return to original directory
cd ..
```

## Test Runner Script

Create a script to run the tests:

```bash
#!/bin/bash
# scripts/run-tests.sh

# Build the service Docker image first
./scripts/build-service-image.sh

# Copy WireMock stubs from the service repository
if [ -d "mq-rest-sync-adapter/src/test/resources/wiremock/rest-api-gateway/mappings" ]; then
  echo "Copying WireMock mappings from service repository..."
  mkdir -p wiremock/mappings
  cp -r mq-rest-sync-adapter/src/test/resources/wiremock/rest-api-gateway/mappings/* wiremock/mappings/
fi

if [ -d "mq-rest-sync-adapter/src/test/resources/wiremock/rest-api-gateway/__files" ]; then
  echo "Copying WireMock files from service repository..."
  mkdir -p wiremock/__files
  cp -r mq-rest-sync-adapter/src/test/resources/wiremock/rest-api-gateway/__files/* wiremock/__files/
fi

# Start the test environment
docker-compose up -d

# Wait for services to be fully available
echo "Waiting for services to start..."
sleep 10

# Run the Cucumber tests
./gradlew cucumber

# Capture the test result
TEST_RESULT=$?

# Stop and remove containers
docker-compose down

# Return the test result
exit $TEST_RESULT
```

## CI Integration

Add a `.gitlab-ci.yml` file to integrate with GitLab CI/CD:

```yaml
stages:
  - test

variables:
  DOCKER_HOST: tcp://docker:2375
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: ""

component-tests:
  stage: test
  image: gradle:jdk17
  services:
    - docker:dind
  script:
    - chmod +x scripts/build-service-image.sh
    - chmod +x scripts/run-tests.sh
    - export MR_ID=$CI_MERGE_REQUEST_IID
    - ./scripts/run-tests.sh
  artifacts:
    paths:
      - build/reports/cucumber/report.html
    when: always
  only:
    - merge_requests
```

## Running Tests Locally

To run the component tests locally:

1. Make scripts executable:
   ```bash
   chmod +x scripts/build-service-image.sh
   chmod +x scripts/run-tests.sh
   ```

2. Run the tests:
   ```bash
   ./scripts/run-tests.sh
   ```

Alternatively, if you want to run the services manually and then execute the tests:

1. Build the service image:
   ```bash
   ./scripts/build-service-image.sh
   ```

2. Start the Docker Compose environment:
   ```bash
   docker-compose up -d
   ```

3. Run the Cucumber tests:
   ```bash
   ./gradlew cucumber
   ```

4. Stop the environment when done:
   ```bash
   docker-compose down
   ```

## Conclusion

This guide provides a complete setup for implementing component tests for the MQ-REST-Sync-Adapter service using Docker Compose, WireMock, and Cucumber Gherkin.

Key aspects of this setup:
- Docker Compose for managing test infrastructure including IBM MQ, WireMock REST API stub, and the service itself
- WireMock for mocking the REST API dependency