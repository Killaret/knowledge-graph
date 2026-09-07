@auth-real @regression
Feature: User authentication
  As a user
  I want to log in and register
  So that I can securely access and manage my notes

  Background:
    Given the test user exists
    And I start an anonymous session

  Scenario: Log in with valid credentials
    Given I am on the login page
    When I enter "testuser" and "TestPassword123!"
    And I click the sign in button
    Then I should be redirected to the main page

  Scenario: Log in with invalid credentials shows an error
    Given I am on the login page
    When I enter "testuser" and "WrongPassword123!"
    And I click the sign in button
    Then I should see an authentication error

  Scenario: Register a new user
    Given I am on the registration page
    When I fill in the registration form with a unique login and password "BddPassword123!"
    And I click the register button
    Then I should be redirected to the main page

  Scenario: Accessing a protected page while logged out redirects to login
    When I navigate to "/profile"
    Then I should be redirected to "/auth/login"
