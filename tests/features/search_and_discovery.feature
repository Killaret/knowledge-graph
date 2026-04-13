Feature: Search and Discovery
  As a user
  I want to search and discover related content
  So that I can find relevant notes quickly

  Background:
    Given the application is open on the graph view
    And various notes exist in the system

  Scenario: Quick search from graph view
    Given I am on the graph view
    When I click the search icon in floating controls
    Then a search input appears
    When I type "quantum"
    Then a dropdown with matching notes appears
    And each result shows the note title and snippet

  Scenario: Navigate to search result
    Given I have searched for "physics"
    And search results are displayed
    When I click on a result "Quantum Physics"
    Then the graph centers on the node "Quantum Physics"
    And the node pulses to indicate location
    And the side panel opens with note details

  Scenario: Full text search
    Given notes with various content exist
    When I search for "relativity"
    Then notes containing "relativity" in title or content are found
    And the matching text is highlighted in results

  Scenario: Semantic search
    Given semantic search is enabled
    When I search for "space exploration"
    Then notes about "rockets", "mars missions", and "astronauts" are also found
    And results are ranked by relevance score

  Scenario: Empty search results
    Given I search for a term with no matches
    When I type "xyznonexistent"
    Then a "No results found" message is displayed
    And suggestions for similar terms are shown

  Scenario: Search history
    Given I have performed searches before
    When I click on the search input
    Then my recent searches are displayed
    When I click on a recent search
    Then the search is executed again

  Scenario: Filter search results
    Given I have searched for "star"
    When I click the "Star" type filter
    Then only "Star" type notes matching "star" are shown
    When I clear the filter
    Then all notes matching "star" are shown

  Scenario: Related notes discovery
    Given I am viewing a note "Solar System"
    When I look at the side panel
    Then a "Related Notes" section shows semantically similar notes
    And related notes are sorted by similarity score

  Scenario: Graph path finder
    Given notes "A" and "Z" exist with multiple paths between them
    When I select node "A" and then Shift+click node "Z"
    Then the shortest path between "A" and "Z" is highlighted
    And intermediate nodes on the path are emphasized
