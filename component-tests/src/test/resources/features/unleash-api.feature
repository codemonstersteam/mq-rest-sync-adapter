@test-unleash @test-stubs @project-setup
Feature: Unleash Project and Feature Toggle Setup

  Background:
    Given Unleash server is running
    And I have default admin API token

  @test-unleash @create-project
  Scenario: Create mq-rest-sync-adapter project
    Given Project with name "default" exists
    Then I create a feature toggle "wallet-balance-workflow" in the project with type "experiment"
    Then the feature toggle should be created successfully in the project
    When I enable the feature toggle in environment "development"
    Then the feature toggle should be enabled in environment "development"
    When I disable the feature toggle in environment "development"
    Then the feature toggle should be disabled in environment "development"
    
