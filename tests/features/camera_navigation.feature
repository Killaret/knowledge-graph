Feature: Camera Position and Navigation in 3D Graph
  As a user
  I want the camera to be positioned correctly when viewing the 3D graph
  So that I can see the relevant notes clearly in different modes

  Background:
    Given the application backend is running
    And the frontend is accessible at "http://localhost:5173"

  Scenario: Camera centers on start node in local graph view
    Given a note "Center Node" exists with type "star"
    And a note "Linked Node" exists with type "planet"
    And a link exists from "Center Node" to "Linked Node"
    When I navigate to the 3D graph page for "Center Node"
    Then the graph renders within 4 seconds
    And the camera is positioned to show "Center Node" in the center
    And the camera distance allows viewing both connected nodes

  Scenario: Camera shows all nodes in full 3D graph view
    Given multiple notes of different types exist
    And links connect some of the notes
    When I navigate to "/graph/3d"
    Then the full graph renders within 5 seconds
    And the camera is positioned to show all visible nodes
    And the camera target is at the center of the node cluster

  Scenario: Camera adjusts when toggling between local and full graph
    Given I am viewing the local 3D graph for a specific note
    And the graph is fully loaded
    When I click the "Show all notes" toggle
    Then the graph switches to full graph mode
    And the camera smoothly animates to show all nodes
    And no node is outside the camera view
    When I click the toggle again
    Then the camera returns to local view positioning

  Scenario: Camera maintains position on browser back/forward
    Given I am viewing the 3D graph for a note
    And the camera is positioned at a specific angle
    When I navigate to the home page
    And I click the browser back button
    Then I return to the 3D graph page
    And the graph renders correctly

  Scenario: Transition from 2D graph to 3D graph maintains context
    Given I am viewing the 2D graph at "/graph"
    And the 2D graph shows nodes
    When I navigate to the 3D graph for the same note
    Then the 3D graph loads
    And the same nodes are visible
    And the camera shows the relevant constellation

  Scenario: Camera handles empty graph gracefully
    Given no notes exist in the system
    When I navigate to "/graph/3d"
    Then an empty state is displayed
    And the camera does not throw errors
    And the scene background is visible

  Scenario: Camera positions correctly for isolated single node
    Given a note "Lonely Node" exists with no connections
    When I navigate to the 3D graph for "Lonely Node"
    Then the graph renders the single node
    And the camera centers on the node
    And the node is clearly visible without other nodes blocking it

  Scenario: Full 3D graph accessible from home page button
    Given I am on the home page
    And notes exist in the system
    When I click the "3D" button in the view controls
    Then I am navigated to "/graph/3d"
    And the full 3D graph loads
    And the camera shows all notes from an appropriate distance

  Scenario: Camera zooms to fit different graph sizes
    Given a small graph with 2-3 nodes exists
    When I view it in 3D mode
    Then the camera distance is appropriate for the small graph
    Given a large graph with 20+ nodes exists
    When I view it in 3D mode
    Then the camera distance is increased to show all nodes
    And no nodes are clipped by the camera frustum

  Scenario: Direct URL access to 3D graph works correctly
    Given I have the URL for a specific note's 3D graph
    When I enter the URL directly in the browser
    Then the 3D graph page loads
    And the camera positions for local view
    And the stats show "Local view" mode
    Given I have the URL for the full 3D graph
    When I enter the URL directly in the browser
    Then the full 3D graph page loads
    And the camera positions for full view
    And the stats show "Full graph" mode
