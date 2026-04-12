Feature: Import and Export
  As a user
  I want to import external documents and export my knowledge graph

  Scenario: Import from a file
    Given I am on the graph view
    When I open the menu and select "Import"
    And I choose a Markdown file "notes.md"
    And I start the import
    Then a progress indicator is shown
    And after completion, new nodes appear on the graph

  Scenario: Export to JSON
    Given I have several notes on the graph
    When I open the menu and select "Export"
    And I choose format "JSON"
    Then a file "knowledge_graph_export.json" is downloaded

  Scenario: Export to CSV
    Given I have several notes on the graph
    When I open the menu and select "Export"
    And I choose format "CSV"
    Then a file "knowledge_graph_export.csv" is downloaded
