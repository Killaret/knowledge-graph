Feature: Note Management
  As a user
  I want to create, edit, and organize my notes
  So that I can build my knowledge graph

  Background:
    Given the application is open
    And I am logged in as a registered user

  Scenario: Create note with wiki links
    Given I am creating a new note
    When I fill in "Title" with "Cosmos Overview"
    And I fill in "Content" with "The cosmos is vast. See also [[Galaxy Formation]]."
    And I select type "Galaxy"
    And I click "Save"
    Then the note is saved
    And a node "Cosmos Overview" appears on the graph
    And a link to "Galaxy Formation" is created if it exists

  Scenario: Create note with new wiki link creates placeholder
    Given I am creating a new note
    When I fill in "Content" with "Related to [[New Topic]]"
    And I click "Save"
    Then the note is saved
    And a placeholder node "New Topic" appears on the graph
    And the placeholder is visually distinct (dashed border)

  Scenario: Edit note content
    Given a note "Existing Note" exists
    When I open the note for editing
    And I change the content to "Updated content with [[Another Link]]"
    And I click "Save"
    Then the note content is updated
    And new links are created based on wiki syntax

  Scenario: Change note type
    Given a note "My Note" of type "Planet" exists
    When I edit the note
    And I change the type to "Star"
    And I click "Save"
    Then the node appearance changes to reflect the new type
    And the node size increases (stars are larger than planets)

  Scenario: Filter notes by type
    Given notes of types "Star", "Planet", and "Comet" exist
    When I click the "Star" filter button
    Then only "Star" type nodes are visible on the graph
    When I click the "All" filter button
    Then all nodes are visible again

  Scenario: Sort notes alphabetically
    Given notes "Zebra", "Apple", and "Mango" exist
    When I select sort option "Alphabetical (A-Z)"
    Then the notes are ordered "Apple", "Mango", "Zebra"

  Scenario: Sort notes by creation date
    Given notes created on different dates exist
    When I select sort option "Newest first"
    Then the most recently created note appears first

  Scenario: Batch select and delete
    Given multiple notes exist on the graph
    When I hold Shift and click on 3 nodes
    Then the 3 nodes are selected (highlighted)
    When I press Delete key
    Then a confirmation modal appears
    When I confirm
    Then all 3 selected nodes are deleted
