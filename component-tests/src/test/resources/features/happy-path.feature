@test-sut
Feature: Happy Path Message Flow
  As a system
  I want to verify the complete message flow through the system
  So that I can ensure all components work together correctly
  
  Background:
    Given Unleash server is running
    And I have default admin API token  

  Scenario: Complete message flow for wallet balance request
    Given MQ server is online
    Given Project with name "default" exists
    Then I create a feature toggle "wallet-balance-workflow" in the project with type "experiment"
    Then the feature toggle should be created successfully in the project
    When I enable the feature toggle in environment "development"
    Then the feature toggle should be enabled in environment "development"
    Then I wait 10 seconds
    And Create valid message payload from "messages/success/request.json"
    When I send the message to "IN.QUEUE.PS"
    Then I read the message from RECEIPT queue "RECEIPT.QUEUE.PS"
    Then I read the message from RECEIPT queue "RECEIPT.QUEUE.PS"
    Then I check that i have 2 receipts with flag SUCCESS and ACCEPTED
    Then I read the message from response queue "OUT.QUEUE.PS"
    When I disable the feature toggle in environment "development"
    Then the feature toggle should be disabled in environment "development"
    And Create valid message payload from "messages/success/request.json"
    When I send the message to "IN.QUEUE.PS"
    Then I read the message from RECEIPT queue "RECEIPT.QUEUE.PS"
    Then I read the message from RECEIPT queue "RECEIPT.QUEUE.PS"
    Then I check that i have 2 receipts with flag SUCCESS and ACCEPTED
    Then I read the message from response queue "OUT.QUEUE.PS"
    