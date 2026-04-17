@smoke @3d-graph @interaction
Feature: 3D Graph Interaction
  As a user
  I want to interact with the graph nodes and camera
  So that I can explore the knowledge graph intuitively

  Background:
    Given I am on the 3D graph page for a note with connections
    And the graph has fully loaded

  @regression
  Scenario: Click on a node opens side panel
    When I click on a visible node
    Then a side panel should open
    And the panel should display the note title
    And the panel should contain a link to the note details

  @regression
  Scenario: Camera controls work during loading
    Given the graph is still loading
    When I drag to rotate the camera
    Then the view should change accordingly
    And the loading overlay should not block the interaction

  @regression
  Scenario: Camera reset button zooms to fit all nodes
    When I click the "Reset Camera" button
    Then the camera should smoothly animate to show all nodes
    And the animation should complete within 1 second

  @regression
  Scenario: Hover over node highlights it
    When I hover over a node
    Then the node should visually highlight
    And a tooltip should appear with the note title

  @regression
  Scenario: Double-click on node navigates to its local graph
    When I double-click on a node
    Then I should navigate to "/graph/3d/{nodeId}"
    And the new graph should center on that node
