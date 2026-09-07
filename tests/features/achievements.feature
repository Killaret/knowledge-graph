Feature: Achievements System
  As a user
  I want to unlock achievements through my actions
  So that I can track my progress and be motivated to engage with the system

  Background:
    Given the application is running
    And I am logged in as a user

  @smoke @achievements
  Scenario: User can view all available achievements
    When I navigate to the achievements page
    Then I should see a list of achievements
    And each achievement should have a title
    And each achievement should have a description
    And each achievement should have an icon

  @achievements
  Scenario: User can view their unlocked achievements
    Given I have unlocked some achievements
    When I navigate to the achievements page
    Then I should see my unlocked achievements marked as earned
    And I should see the unlock date for each earned achievement

  @achievements
  Scenario: User receives notification for new achievement
    Given I have not unlocked the "First Note" achievement
    When I create a new note
    Then the "First Note" achievement should be unlocked
    And I should see a toast notification about the achievement
    And the notification should use the galactic lexicon if galactic mode is enabled

  @achievements
  Scenario: Achievement notification can be dismissed
    Given I have a new achievement notification
    When I dismiss the notification
    Then the notification should be marked as seen
    And the notification should not appear again
    And the API should be called to mark it as seen

  @achievements @galactic
  Scenario: Achievement notification respects galactic mode setting
    Given I have galactic mode enabled
    When I unlock an achievement
    Then the notification should use galactic-themed messages
    And the notification should display space-themed metaphors

  @achievements @galactic
  Scenario: Achievement notification respects standard mode setting
    Given I have galactic mode disabled
    When I unlock an achievement
    Then the notification should use standard technical messages
    And the notification should not use space-themed metaphors

  @achievements
  Scenario: Achievement notifications can be disabled
    Given I have achievement notifications disabled in settings
    When I unlock an achievement
    Then I should not see a toast notification
    But the achievement should still be unlocked in the system

  @achievements
  Scenario: Achievement triggers on note creation
    Given I have 0 notes created
    And there is an achievement for creating 1 note
    When I create a new note
    Then the achievement should be unlocked
    And the achievement should be added to my achievements list

  @achievements
  Scenario: Achievement triggers on link creation
    Given I have 0 links created
    And there is an achievement for creating 1 link
    When I create a new link between notes
    Then the achievement should be unlocked
    And the achievement should be added to my achievements list

  @achievements
  Scenario: Achievement with threshold requires multiple actions
    Given I have 4 notes created
    And there is an achievement for creating 5 notes
    When I create a new note
    Then the achievement should be unlocked
    And I should have 5 notes total

  @achievements
  Scenario: Achievement with type filter
    Given I have 2 star notes created
    And there is an achievement for creating 3 star notes
    When I create a new star note
    Then the achievement should be unlocked
    And the achievement should only count star notes

  @achievements
  Scenario: Hidden achievements are not shown until unlocked
    Given there is a hidden achievement
    When I navigate to the achievements page
    Then I should not see the hidden achievement
    When I unlock the hidden achievement
    Then I should see the hidden achievement on the achievements page

  @achievements
  Scenario: Achievement polling updates new achievements
    Given I have unlocked a new achievement on the server
    And the frontend is polling for achievements
    When the polling interval elapses
    Then the new achievement should appear in my achievements list
    And I should see a notification for the new achievement
