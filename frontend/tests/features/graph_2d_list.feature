@smoke @2d-graph @list-view
Feature: 2D Graph and List View
  As a user
  I want to toggle between graph and list views
  So that I can explore notes in different ways

  Background:
    Given I am on the main page "/"
    And there are notes of various types in the database

  @regression
  Scenario: Toggle between graph and list views
    Then I should see the 2D force graph by default
    When I click the "List" toggle button in the floating controls
    Then I should see a grid of note cards
    And the view toggle should show "Graph" option
    When I click the "Graph" toggle button
    Then I should see the fullscreen 2D force graph
    And the graph canvas should be visible

  @regression
  Scenario: Filter notes by type in list view
    Given I am in list view
    When I click the "Planet" filter chip in floating controls
    Then only notes of type "Planet" should be displayed
    And the count badge should show the correct number
    When I click the "All" filter chip
    Then all notes should be displayed

  @regression
  Scenario: Search filters notes in list view
    Given I am in list view
    When I type "Test star" in the search input
    Then the list should show only notes containing "Test star"
    And the note cards should highlight the matching text
    When I clear the search input
    Then all notes should be displayed again

  @regression
  Scenario: Search filters nodes in 2D graph view
    Given I am in graph view
    When I type "star" in the search input
    Then only nodes matching "star" should be visible
    And non-matching nodes should be dimmed or hidden
    When I clear the search input
    Then all nodes should be visible

  @regression
  Scenario: Create note from floating controls
    Given I am on the main page
    When I click the "+" button in floating controls
    Then a create note modal should open
    When I fill in the title "Test Note"
    And I select type "Star"
    And I click the "Create" button
    Then the modal should close
    And the new note should appear in the graph
