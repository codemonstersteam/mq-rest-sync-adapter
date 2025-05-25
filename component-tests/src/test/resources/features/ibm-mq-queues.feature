@test-stubs
Feature: IBM MQ Queue Operations
  As a system
  I want to verify that messages can be sent to and received from IBM MQ queues
  So that I can ensure the queue system is working correctly for the component tests

  # Тест нужен чтобы отработал тест в исходном сервисе (есть сомнение что он нужен)
  Scenario: Test simple message in MY.TEST queue
    Given IBM MQ server is running and accessible
    And I have a valid request payload from "simple-test-payload.json"
    When I send the message to queue "MY.TEST"
    Then the message should be successfully sent
    When I read the message from queue "MY.TEST"
    Then the received message should match the sent payload 
  
  Scenario: Send message to IN.QUEUE.PS and verify it can be read back
    Given IBM MQ server is running and accessible
    And I have a valid request payload from "request-payload.json"
    When I send the message to queue "IN.QUEUE.PS"
    Then the message should be successfully sent
    When I read the message from queue "IN.QUEUE.PS"
    Then the received message should match the sent payload
    And the message should contain the correct correlation ID
    And the message should contain the correct operation type 

  Scenario: Test simple message in OUT.QUEUE.PS queue
    Given IBM MQ server is running and accessible
    And I have a valid request payload from "simple-test-payload.json"
    When I send the message to queue "OUT.QUEUE.PS"
    Then the message should be successfully sent
    When I read the message from queue "OUT.QUEUE.PS"
    Then the received message should match the sent payload 
  
  Scenario: Test simple message in RECEIPT.QUEUE.PS queue
    Given IBM MQ server is running and accessible
    And I have a valid request payload from "simple-test-payload.json"
    When I send the message to queue "RECEIPT.QUEUE.PS"
    Then the message should be successfully sent
    When I read the message from queue "RECEIPT.QUEUE.PS"
    Then the received message should match the sent payload 
