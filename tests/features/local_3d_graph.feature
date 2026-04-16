Feature: Local 3D Graph View
  As a user
  I want to view a specific note and its connections in 3D
  So that I can focus on a particular part of my knowledge graph

  Background:
    Given the application backend is running
    And the frontend is accessible at "http://localhost:5173"

  Scenario: Navigate to local 3D view from note detail page
    Given a note "Central Note" exists with content "Main content"
    And note "Central Note" is linked to note "Related Note 1"
    And note "Central Note" is linked to note "Related Note 2"
    When I am on the note detail page for "Central Note"
    And I click the "3D View" button
    Then I am navigated to "/graph/3d/{noteId}"
    And the local 3D graph page loads
    And the stats bar shows "Local view" mode
    And the 3D graph shows "Central Note" as the central node
    And the graph shows connections to related notes

  Scenario: Navigate to local 3D view from home page with selected note
    Given I am on the home page
    And multiple notes exist on the graph
    When I click on node "Target Note"
    Then the side panel opens with note details
    When I click the "3D" button in floating controls
    Then I am navigated to "/graph/3d/{targetNoteId}"
    And the local 3D graph renders with "Target Note" as center
    And the stats bar shows "Local view" mode

  Scenario: Local 3D view shows single note with fog when no connections
    Given a note "Isolated Note" exists with content "No links"
    When I navigate to "/graph/3d/{isolatedNoteId}"
    Then the local 3D graph renders
    And the graph shows only "Isolated Note" node
    And the fog effect starts dense
    And the fog gradually dissipates
    And the single node is centered in camera view

  Scenario: Local 3D view with progressive loading
    Given a note "Hub Note" exists with content "Connected to many"
    And "Hub Note" has 10 related notes
    When I navigate to "/graph/3d/{hubNoteId}"
    Then the graph appears immediately without spinner
    And the stats bar shows initial node count
    And the fog effect is active during loading
    When I wait for progressive loading to complete
    Then the stats bar updates with full node count
    And the fog gradually dissipates
    And the camera centers on all loaded nodes

  Scenario: Switch from local to full graph view using toggle
    Given I am on the local 3D graph view for note "Test Note"
    And "Test Note" has connections to other notes
    When I click the "Show all notes" toggle
    Then the stats bar updates to show "Full graph" mode
    And the 3D graph updates to show all notes in the system
    And the fog animation plays during transition
    And the camera re-centers on all visible nodes

  Scenario: Switch back from full to local graph view
    Given I am on the full 3D graph at "/graph/3d"
    When I navigate to "/graph/3d/{noteId}"
    Then the local 3D graph renders
    And the stats bar shows "Local view" mode
    And only notes connected to the selected note are visible
    And the camera centers on the local graph cluster

  Scenario: Local 3D view camera behavior
    Given a note "Camera Test" exists with 5 related notes
    When I navigate to "/graph/3d/{cameraTestId}"
    Then the 3D graph renders
    And the camera is positioned to show all nodes in the local graph
    And the camera maintains minimum distance of 30 units
    And all local nodes are within camera view

  Scenario: Navigate from local 3D to another note's local view
    Given I am on the local 3D graph view for note "Note A"
    And "Note A" is connected to "Note B"
    When I click on node "Note B" in the 3D graph
    Then the side panel opens with "Note B" details
    When I navigate to "/graph/3d/{noteBId}"
    Then the local 3D graph renders with "Note B" as center
    And the stats bar shows "Local view" mode

  Scenario: Links remain visible when switching from local to full graph
    Given a note "Hub Note" exists with content "Central node"
    And "Hub Note" has 3 related notes
    When I navigate to "/graph/3d/{hubNoteId}"
    Then the local 3D graph renders with "Hub Note" as center
    And the graph shows connections to related notes
    And the stats bar shows at least 4 nodes
    When I click the "Show all notes" toggle
    Then the stats bar updates to show "Full graph" mode
    And the 3D graph renders
    And the stats bar shows link count greater than 0
    And all existing links remain visible without flickering

  Scenario: Links persist during camera zoom operations
    Given a note "Zoom Test" exists with content "Testing zoom"
    And "Zoom Test" has 2 related notes
    When I navigate to "/graph/3d/{zoomTestId}"
    Then the local 3D graph renders with "Zoom Test" as center
    And the graph shows connections to related notes
    When I zoom in on the graph
    Then the links remain connected to their nodes
    And no links appear disconnected or floating
    When I zoom out on the graph
    Then the links still connect the same nodes
    And the link count in stats bar remains constant

  Scenario: Links persist during camera rotation
    Given a note "Rotate Test" exists with content "Testing rotation"
    And "Rotate Test" has 2 related notes
    When I navigate to "/graph/3d/{rotateTestId}"
    Then the local 3D graph renders with "Rotate Test" as center
    And the graph shows connections to related notes
    When I rotate the camera 90 degrees around the graph
    Then the links remain connected to their nodes
    And the links rotate with the nodes
    And no links are lost during rotation
    When I rotate the camera another 90 degrees
    Then all links are still visible and connected

  Scenario: Links are not duplicated when switching views multiple times
    Given a note "Switch Test" exists with content "Testing switches"
    And "Switch Test" has 2 related notes
    When I navigate to "/graph/3d/{switchTestId}"
    Then the local 3D graph renders with "Switch Test" as center
    And the stats bar shows at least 3 nodes
    And I record the initial link count
    When I click the "Show all notes" toggle
    Then the stats bar updates to show "Full graph" mode
    When I click the "Show all notes" toggle again
    Then the stats bar updates to show "Local view" mode
    And the link count matches the initial recorded count
    And no duplicate links are present in the graph

  Scenario: Links connect correct nodes after progressive loading completes
    Given a note "Progressive Test" exists with content "Testing progressive"
    And "Progressive Test" has 5 related notes
    When I navigate to "/graph/3d/{progressiveTestId}"
    Then the local 3D graph renders with "Progressive Test" as center
    And the stats bar shows initial node count
    And I note which nodes are connected
    When I wait for progressive loading to complete
    Then the stats bar updates with full node count
    And all previously noted connections remain intact
    And new connections are properly linked to correct nodes
