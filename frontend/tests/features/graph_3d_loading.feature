@smoke @3d-graph
Feature: 3D Graph Progressive Loading
  As a user
  I want to see the graph appearing gradually with fog
  So that I can start interacting early

  Background:
    Given I have test notes with connections
    And I navigate to "/graph/3d/{centerNoteId}"

  @regression
  Scenario: Overlay disappears early while fog persists
    Then the loading overlay should be visible
    And the loading text should contain "universe isn't still created"
    And the fog density should be at least 0.005
    When the simulation progresses to at least 10% nodes positioned
    Then the loading overlay should disappear within 2 seconds
    But the fog density should still be greater than 0.005
    And I should be able to click on the canvas

  @regression
  Scenario: Fog fully clears after simulation stabilizes
    When I wait for the simulation to end
    Then the fog density should become less than 0.01 within 1 second
    And all nodes should be clickable
    And links should be visible with opacity based on weight

  @regression
  Scenario: Early interaction is not blocked by overlay
    Given the graph is still loading
    When I drag the canvas to rotate the view
    Then the camera position should change
    And the loading overlay should still be visible or fading
