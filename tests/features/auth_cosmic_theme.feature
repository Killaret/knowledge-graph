Feature: Authentication Pages with Cosmic Theme
  As a user
  I want to see beautifully designed authentication pages
  So that the login experience feels immersive and matches the cosmic theme

  Background:
    Given the application is running

  @smoke @auth @visual
  Scenario: Login page displays cosmic background
    When I navigate to the login page
    Then I should see a cosmic starfield background
    And the background should have animated stars

  @smoke @auth @visual
  Scenario: Login page displays galaxy icon
    When I navigate to the login page
    Then I should see a rotating galaxy icon above the title
    And the icon should have a glowing effect

  @smoke @auth @visual
  Scenario: Login form has glass morphism card
    When I navigate to the login page
    Then I should see a login form in a glass-like card
    And the card should have a backdrop blur effect
    And the card should have a subtle golden border glow

  @auth @visual
  Scenario: Login inputs have cosmic focus effect
    When I navigate to the login page
    And I click on the login input field
    Then the input should have a golden glow border
    And the glow should animate smoothly

  @smoke @auth
  Scenario: Login form is functional
    When I navigate to the login page
    And I enter "testuser" in the login field
    And I enter "testpassword123!" in the password field
    Then the login button should be enabled

  @smoke @auth @visual
  Scenario: Register page displays cosmic theme
    When I navigate to the register page
    Then I should see a cosmic starfield background
    And I should see a galaxy icon
    And I should see the title "Create Account"

  @auth
  Scenario: Register form validates password requirements
    When I navigate to the register page
    And I enter "weak" in the password field
    Then I should see password requirements list
    And the requirements should show which are not met
    When I enter a valid password "StrongPass123!"
    Then all requirements should be marked as valid

  @smoke @auth @visual
  Scenario: Forgot password page displays cosmic theme
    When I navigate to the forgot password page
    Then I should see a cosmic starfield background
    And I should see the title "Password Recovery"

  @smoke @auth @visual
  Scenario: Reset password page displays cosmic theme with token
    When I navigate to the reset password page with a valid token
    Then I should see a cosmic starfield background
    And I should see the title "Reset Password"
    And I should see password input fields

  @auth @visual
  Scenario: Reset password page shows error without token
    When I navigate to the reset password page without a token
    Then I should see an error message
    And I should see a constellation icon
    And I should see a link to request a new reset

  @auth @visual @animation
  Scenario: Auth card has entrance animation
    When I navigate to the login page
    Then the auth card should animate in with a fly effect
    And the animation should fade in smoothly

  @auth @visual
  Scenario: All auth pages have consistent styling
    When I navigate to the login page
    Then the page should have cosmic theme styling
    When I navigate to the register page
    Then the page should have the same cosmic theme styling
    When I navigate to the forgot password page
    Then the page should have the same cosmic theme styling

  @auth @yandex
  Scenario: Yandex login button has cosmic hover effect
    When I navigate to the login page
    And I hover over the Yandex login button
    Then the button should have a glow effect
    And the button should lift slightly

  @auth @performance
  Scenario: Cosmic background does not impact performance
    When I navigate to the login page
    Then the starfield animation should run at 60fps
    And CPU usage should remain low
