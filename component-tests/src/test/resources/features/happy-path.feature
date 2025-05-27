@test-sut
Feature: Happy Path Message Flow
  As a system
  I want to verify the complete message flow through the system
  So that I can ensure all components work together correctly

  Scenario: Complete message flow for wallet balance request
    Given MQ server is accessible
    And Create valid message payload from "messages/success/request.json"
    When I send the message to IN queue "IN.QUEUE.PS"
    Then I read the message from RECEIPT queue "RECEIPT.QUEUE.PS" with flag "ACCEPTED"
    Then I read the message from RECEIPT queue "RECEIPT.QUEUE.PS" with flag "SUCCESS"
    Then I read the message from response queue "OUT.QUEUE.PS"
