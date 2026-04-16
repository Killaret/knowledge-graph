Feature: Link Types and Visualization
  As a user
  I want to see different types of links between notes
  So that I can understand the nature of relationships in my knowledge graph

  Background:
    Given the application backend is running
    And multiple notes exist on the graph

  Scenario: Reference link is rendered with solid line
    Given notes "Source Note" and "Target Note" exist
    And a link exists from "Source Note" to "Target Note" with type "reference"
    When I navigate to the 3D graph page for "Source Note"
    Then the link to "Target Note" is visible
    And the link has solid line style

  Scenario: Dependency link is rendered with dashed line
    Given notes "Source Note" and "Target Note" exist
    And a link exists from "Source Note" to "Target Note" with type "dependency"
    When I navigate to the 3D graph page for "Source Note"
    Then the link to "Target Note" is visible
    And the link has dashed line style

  Scenario: Related link is rendered with dotted line
    Given notes "Source Note" and "Target Note" exist
    And a link exists from "Source Note" to "Target Note" with type "related"
    When I navigate to the 3D graph page for "Source Note"
    Then the link to "Target Note" is visible
    And the link has dotted line style

  Scenario: Custom link type is rendered with default styling
    Given notes "Source Note" and "Target Note" exist
    And a link exists from "Source Note" to "Target Note" with type "custom_type"
    When I navigate to the 3D graph page for "Source Note"
    Then the link to "Target Note" is visible
    And the link uses default styling

  Scenario: Strong link has thicker line
    Given notes "Source Note" and "Target Note" exist
    And a link exists from "Source Note" to "Target Note" with weight 0.9
    When I navigate to the 3D graph page for "Source Note"
    Then the link to "Target Note" has thick line width

  Scenario: Weak link has thinner line
    Given notes "Source Note" and "Target Note" exist
    And a link exists from "Source Note" to "Target Note" with weight 0.2
    When I navigate to the 3D graph page for "Source Note"
    Then the link to "Target Note" has thin line width

  Scenario: Multiple link types are visible simultaneously
    Given note "Hub Note" exists
    And note "Reference Target" is linked with type "reference"
    And note "Dependency Target" is linked with type "dependency"
    And note "Related Target" is linked with type "related"
    When I navigate to the 3D graph page for "Hub Note"
    Then all 3 links are visible
    And each link has distinct visual styling

  Scenario: Link connects nodes even when start node not in API response
    Given note "Start Node" has no connections in API
    But note "Start Node" is the center of local graph view
    When I navigate to the 3D graph page for "Start Node"
    Then the graph shows "Start Node" in the center
    And stats show 1 node and 0 links
    And no error is displayed
