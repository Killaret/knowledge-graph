Feature: Graph View Modes
  As a user
  I want to switch between 2D and 3D graph views
  So that I can choose between performance and visual appeal

  Background:
    Given the application is open on the graph view
    And multiple notes exist on the graph

  Scenario: 2D graph is displayed by default
    Given I am on the main page
    Then the graph canvas shows "2D Mode" indicator
    And nodes are rendered as 2D shapes

  Scenario: Switch to 3D view
    Given I am on the graph view in 2D mode
    When I click the "3D View" button in floating controls
    Then the graph switches to 3D mode
    And the graph canvas shows "3D Mode" indicator
    And nodes are rendered as 3D spheres

  Scenario: Return to 2D view from 3D
    Given I am on the graph view in 3D mode
    When I click the "2D View" button
    Then the graph switches back to 2D mode
    And the graph canvas shows "2D Mode" indicator

  Scenario: 3D view on low-end devices
    Given I am using a low-end device
    When I open the graph view
    Then the graph defaults to 2D mode
    And the "3D View" button shows a warning icon
    And clicking it shows a performance warning

  Scenario: Zoom and pan in 2D mode
    Given I am on the graph view in 2D mode
    When I scroll the mouse wheel up
    Then the graph zooms in
    When I scroll the mouse wheel down
    Then the graph zooms out
    When I drag the canvas
    Then the graph pans in the dragged direction

  Scenario: Rotate view in 3D mode
    Given I am on the graph view in 3D mode
    When I drag the mouse on the canvas
    Then the camera rotates around the graph
    When I use the scroll wheel
    Then the camera zooms in and out

  Scenario: Click node to open details in 2D
    Given I am on the graph view in 2D mode
    When I click on a node
    Then a side panel opens showing note details
    And the node is highlighted with a glow effect

  Scenario: Click node to open details in 3D
    Given I am on the graph view in 3D mode
    When I click on a node
    Then a side panel opens showing note details
    And the camera centers on the clicked node
