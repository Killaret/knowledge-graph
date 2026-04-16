Feature: Type Filtering on Home Page
  As a user
  I want to filter notes by their celestial body type
  So that I can focus on specific categories of knowledge

  Background:
    Given the application is open on the home page
    And notes of various types exist in the system

  Scenario: Filter by star type
    Given notes of types "star", "planet", and "comet" exist
    When I click the "Stars" filter button
    Then only star type notes are displayed in the list
    And the stats bar shows "filtered by: Stars"
    And the graph shows only star nodes

  Scenario: Filter by planet type
    Given notes of types "star", "planet", and "comet" exist
    When I click the "Planets" filter button
    Then only planet type notes are displayed in the list
    And the stats bar shows "filtered by: Planets"
    And the graph shows only planet nodes

  Scenario: Filter by comet type
    Given notes of types "star", "planet", and "comet" exist
    When I click the "Comets" filter button
    Then only comet type notes are displayed in the list
    And the stats bar shows "filtered by: Comets"
    And the graph shows only comet nodes

  Scenario: Filter by galaxy type
    Given notes of types "star", "galaxy", and "planet" exist
    When I click the "Galaxies" filter button
    Then only galaxy type notes are displayed in the list
    And the stats bar shows "filtered by: Galaxies"
    And the graph shows only galaxy nodes

  Scenario: Clear filter by selecting All
    Given I have applied the "Stars" filter
    When I click the "All" filter button
    Then all notes are displayed regardless of type
    And the stats bar shows total count without filter indicator
    And the graph shows all nodes

  Scenario: Filter works with type in metadata field
    Given a note exists with metadata type "star" but no root type
    And the filter is currently set to "All"
    When I click the "Stars" filter button
    Then the note with metadata type "star" is included in results
    And it is displayed in the filtered list

  Scenario: Default type falls back to star when not specified
    Given a note exists with no type field specified
    When I click the "Stars" filter button
    Then the note without type is treated as "star" type
    And it appears in the filtered results

  Scenario: Filter state persists when switching views
    Given I have applied the "Planets" filter
    When I switch to list view
    Then the filter remains active
    And only planet notes are shown in the list
    When I switch back to graph view
    Then the filter is still active
    And only planet nodes are shown on the graph

  Scenario: Empty state when filter matches no notes
    Given no notes of type "galaxy" exist
    When I click the "Galaxies" filter button
    Then an empty state message is displayed
    And the message indicates "No galaxies found"
    And the graph shows empty state

  Scenario: Combined search and type filter
    Given notes of various types exist
    And I have typed "cosmos" in the search box
    When I click the "Stars" filter button
    Then only star type notes matching "cosmos" are displayed
    And both search and filter indicators are shown in stats
