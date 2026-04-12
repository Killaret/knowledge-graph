Feature: Import and Export
  As a user
  I want to import external documents and export my knowledge graph

  Background:
    Given the application is open on the graph view

  Scenario: Import from a Markdown file
    Given I am on the graph view
    When I open the menu and select "Import"
    And I choose a Markdown file "notes.md"
    And I start the import
    Then a progress indicator is shown
    And after completion, new nodes appear on the graph

  Scenario: Import from a PDF file
    Given I am on the graph view
    When I open the menu and select "Import"
    And I choose a PDF file "document.pdf"
    And I start the import
    Then a progress indicator is shown
    And after completion, new nodes appear on the graph

  Scenario: Import from URL
    Given I am on the graph view
    When I open the menu and select "Import"
    And I enter URL "https://example.com/article"
    And I start the import
    Then a progress indicator is shown
    And after completion, new nodes appear on the graph

  Scenario: Export to JSON
    Given I have several notes on the graph
    When I open the menu and select "Export"
    And I choose format "JSON"
    Then a file "knowledge_graph_export.json" is downloaded

  Scenario: Export to Markdown
    Given I have several notes on the graph
    When I open the menu and select "Export"
    And I choose format "Markdown"
    Then a file "knowledge_graph_export.md" is downloaded

  Scenario: Export to GraphML
    Given I have several notes on the graph
    When I open the menu and select "Export"
    And I choose format "GraphML"
    Then a file "knowledge_graph_export.graphml" is downloaded
