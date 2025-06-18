@test-sut @rest-resource-unavalable
Feature: Happy Path Message Flow
  As a system
  I want to verify the complete message flow through the system
  While rest resource is unavalable

  Scenario: Test rest resource unavalable
    Given MQ server is accessible
    And Create valid message payload from "messages/success/request.json" for test REST resource unavalable
    When I send the message with header Api-clientId = Api-clientId-02 to in queue "IN.QUEUE.PS"
    Then I read receipt "RECEIPT.QUEUE.PS"
    Then I read receipt "RECEIPT.QUEUE.PS"
    Then I check that i have 2 receipts with flag ACCEPTED and ERROR
    Then I read the error from response queue "OUT.QUEUE.PS"
