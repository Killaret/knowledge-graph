Feature: Celestial Body Types in 3D Graph
  As a user
  I want to see different celestial body types rendered in the 3D graph
  So that I can visually distinguish between different types of notes

  Background:
    Given the application backend is running
    And the frontend is accessible at "http://localhost:5173"

  Scenario: Star type is rendered as glowing sphere with rays
    Given a note exists with title "Star Note" and type "star"
    When I navigate to the 3D graph page for "Star Note"
    Then the 3D graph renders a star celestial body
    And the star has a glowing core
    And the star has radiating rays

  Scenario: Planet type is rendered as sphere with rings
    Given a note exists with title "Planet Note" and type "planet"
    When I navigate to the 3D graph page for "Planet Note"
    Then the 3D graph renders a planet celestial body
    And the planet has ring structures

  Scenario: Comet type is rendered with tail
    Given a note exists with title "Comet Note" and type "comet"
    When I navigate to the 3D graph page for "Comet Note"
    Then the 3D graph renders a comet celestial body
    And the comet has a glowing tail

  Scenario: Galaxy type is rendered as spiral with particles
    Given a note exists with title "Galaxy Note" and type "galaxy"
    When I navigate to the 3D graph page for "Galaxy Note"
    Then the 3D graph renders a galaxy celestial body
    And the galaxy has spiral arm particles

  Scenario: Asteroid type is rendered as irregular rock
    Given a note exists with title "Asteroid Note" and type "asteroid"
    When I navigate to the 3D graph page for "Asteroid Note"
    Then the 3D graph renders an asteroid celestial body
    And the asteroid has irregular rocky surface

  Scenario: Debris type is rendered as scattered particles
    Given a note exists with title "Debris Note" and type "debris"
    When I navigate to the 3D graph page for "Debris Note"
    Then the 3D graph renders a debris field
    And the debris consists of scattered particles

  Scenario: Unknown type falls back to default sphere
    Given a note exists with title "Unknown Note" and type "unknown_type"
    When I navigate to the 3D graph page for "Unknown Note"
    Then the 3D graph renders a default celestial body
    And the default body is a basic sphere

  Scenario: Type from metadata field is used when root type is missing
    Given a note exists with title "Metadata Star" and metadata type "star"
    And the note has no root type field
    When I navigate to the 3D graph page for "Metadata Star"
    Then the 3D graph renders a star celestial body
    And the star is displayed correctly using metadata.type
