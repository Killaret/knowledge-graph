Feature: Graph Navigation
  As a user
  I want to interact with the knowledge graph visually
  So that I can explore and manage my notes intuitively

  Background:
    Given the application is open on the graph view

  Scenario: Empty state prompts note creation
    When there are no notes in the system
    Then the graph canvas shows a message "No notes yet"
    And a "+" button is visible

  Scenario: Create a new note via the plus button
    Given the graph view is active
    When I click the "+" button
    Then a modal with a note form appears
    When I fill in "Title" with "My First Note"
    And I select type "Star"
    And I click "Save"
    Then the modal closes
    And a new node labeled "My First Note" appears on the graph

  Scenario: Edit a note from the graph
    Given a note "My First Note" exists on the graph
    When I click on the node "My First Note"
    Then a side panel opens showing note details
    When I click "Edit" in the side panel
    Then a modal with the note form appears pre-filled
    When I change "Title" to "Updated Note"
    And I click "Save"
    Then the node label changes to "Updated Note"

  Scenario: Delete a note with confirmation
    Given a note "To Be Deleted" exists
    When I click on the node "To Be Deleted"
    And I click "Delete" in the side panel
    Then a confirmation modal appears with text "Delete note?"
    When I confirm the deletion
    Then the node "To Be Deleted" disappears from the graph

  Scenario: Create a link by dragging between nodes
    Given notes "Source" and "Target" exist on the graph
    When I drag from node "Source" to node "Target"
    Then a link preview line is shown
    When I release the mouse over "Target"
    Then a weight selector appears
    When I select weight "4" and confirm
    Then a new edge appears between "Source" and "Target"

  Scenario: Search and focus on a node
    Given multiple notes exist on the graph
    When I click the search icon
    And I type "cosmos" in the search input
    Then a list of matching notes appears
    When I click on a search result "Cosmos Article"
    Then the graph zooms to the node "Cosmos Article"
    And the node is highlighted
