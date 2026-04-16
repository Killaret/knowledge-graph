Feature: Full 3D Graph View
  As a user
  I want to view all my notes in a single 3D graph
  So that I can see the complete structure of my knowledge base

  Background:
    Given the application backend is running
    And the frontend is accessible at "http://localhost:5173"

  Scenario: Navigate to full 3D graph from home page
    Given I am on the home page
    When I click the "3D" button in the view controls
    Then I am navigated to "/graph/3d"
    And the full 3D graph page loads
    And no 404 error is shown

  Scenario: Full 3D graph displays all notes
    Given multiple notes exist in the system
    When I navigate to "/graph/3d"
    Then the 3D graph renders
    And stats show all nodes with count matching total notes
    And stats show all links in the system

  Scenario: Full 3D graph loads without spinner
    Given the full graph has many notes
    When I navigate to "/graph/3d"
    Then the graph appears immediately without spinner
    And progressive loading indicator may appear briefly

  Scenario: Stats bar shows full graph mode
    When I navigate to "/graph/3d"
    And the graph loads successfully
    Then the stats bar shows node and link counts
    And the mode indicator shows "Full graph" or similar

  Scenario: Navigate from full 3D to specific note 3D
    Given I am on the full 3D graph at "/graph/3d"
    And a note "Specific Note" exists
    When I click on node "Specific Note"
    Then the side panel opens with note details
    And I can navigate to the local 3D view for that note

  Scenario: Full 3D graph handles empty database
    Given no notes exist in the system
    When I navigate to "/graph/3d"
    Then an empty state is displayed
    And the message prompts to create first note
    And no errors occur

  Scenario: Full 3D graph with isolated notes
    Given notes exist with no connections between them
    When I navigate to "/graph/3d"
    Then all isolated notes are rendered
    And each appears as separate node
    And stats show correct node count
    And stats show 0 links

  Scenario: Full 3D graph shows fog animation on load
    Given multiple notes exist in the system
    When I navigate to "/graph/3d"
    Then the 3D graph renders
    And the fog effect starts dense
    And the fog gradually dissipates as simulation ends
    And the graph becomes fully visible

  Scenario: Full 3D graph camera centers on all elements
    Given multiple notes exist in the system
    When I navigate to "/graph/3d"
    Then the 3D graph renders
    And the camera is positioned to show all nodes
    And all nodes are within camera view
    And the camera maintains minimum distance of 30 units

  Scenario: Full 3D graph links persist during camera zoom
    Given notes exist with connections between them
    When I navigate to "/graph/3d"
    Then the 3D graph renders
    And the stats bar shows link count greater than 0
    And I record the initial link count
    When I zoom in on the graph
    Then the links remain connected to their nodes
    And the link count matches the initial recorded count
    When I zoom out on the graph
    Then all links are still visible and connected
    And the link count matches the initial recorded count

  Scenario: Full 3D graph links persist during camera rotation
    Given notes exist with connections between them
    When I navigate to "/graph/3d"
    Then the 3D graph renders
    And the stats bar shows link count greater than 0
    When I rotate the camera 90 degrees around the graph
    Then the links remain connected to their nodes
    And no links are lost during rotation
    When I rotate the camera another 90 degrees
    Then all links are still visible and connected

  Scenario: Full 3D graph shows all links correctly with many nodes
    Given the full graph has many notes
    And notes have various connection patterns
    When I navigate to "/graph/3d"
    Then the 3D graph renders
    And the stats bar shows link count greater than 0
    And no duplicate links are present in the graph
    And all existing links remain visible without flickering
