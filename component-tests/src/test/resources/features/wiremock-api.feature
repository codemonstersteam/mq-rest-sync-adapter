@test-stubs
Feature: WireMock API Testing

  Scenario: Test wallet balance API
    Given the WireMock REST API stub is running
    When I send a POST request to "/api/client/Api-clientId-01/wallet/balance"
    Then I should receive a successful response
    And the response should contain Content-Type header "application/json" 
    And the response should contain status "success"
    And the response should contain actualTimestamp "1607449018829"
    And the response should contain data with createTime "1600443296227"
    And the response should contain data with operationId "01"
    And the response should contain balance with value "999.9" and currency "rub"
    