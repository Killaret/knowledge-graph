import requests
import json

BASE_URL = 'http://localhost:8080'

def create_note(title, content, note_type='star'):
    """Create a note and return its ID"""
    resp = requests.post(f'{BASE_URL}/notes', json={
        'title': title,
        'content': content,
        'type': note_type
    })
    if resp.status_code == 201:
        data = resp.json()
        print(f'Created: {title} ({note_type}) - ID: {data["id"]}')
        return data['id']
    else:
        print(f'Failed to create {title}: {resp.status_code}')
        return None

def create_link(source_id, target_id, weight=0.5, link_type='related'):
    """Create a link between notes"""
    resp = requests.post(f'{BASE_URL}/links', json={
        'sourceNoteId': source_id,
        'targetNoteId': target_id,
        'weight': weight,
        'linkType': link_type
    })
    if resp.status_code == 201:
        print(f'  Link: {source_id} -> {target_id} ({link_type}, weight={weight})')
        return True
    else:
        print(f'  Failed link: {resp.status_code}')
        return False

print('=== Creating Test Data ===\n')

# Group 1: Solar System Theme (connected)
print('--- Solar System Group ---')
solar_system = create_note('Solar System', 'Our planetary system with the Sun at its center', 'star')
earth = create_note('Earth', 'Third planet from the Sun, only known planet with life', 'planet')
mars = create_note('Mars', 'Fourth planet, known as the Red Planet', 'planet')
jupiter = create_note('Jupiter', 'Largest planet, gas giant with many moons', 'planet')
asteroid_belt = create_note('Asteroid Belt', 'Region between Mars and Jupiter with many asteroids', 'asteroid')

# Links within solar system
if all([solar_system, earth, mars, jupiter, asteroid_belt]):
    create_link(solar_system, earth, 1.0, 'reference')
    create_link(solar_system, mars, 1.0, 'reference')
    create_link(solar_system, jupiter, 1.0, 'reference')
    create_link(earth, mars, 0.7, 'related')
    create_link(mars, asteroid_belt, 0.8, 'dependency')
    create_link(asteroid_belt, jupiter, 0.6, 'related')

# Group 2: Deep Space Objects (connected)
print('\n--- Deep Space Group ---')
milky_way = create_note('Milky Way', 'Our home galaxy, barred spiral galaxy', 'galaxy')
andromeda = create_note('Andromeda Galaxy', 'Nearest major galaxy to Milky Way', 'galaxy')
black_hole = create_note('Black Hole', 'Region of spacetime with extreme gravitational effects', 'star')

if all([milky_way, andromeda, black_hole]):
    create_link(milky_way, andromeda, 0.9, 'related')
    create_link(milky_way, black_hole, 0.8, 'reference')
    create_link(andromeda, black_hole, 0.5, 'dependency')

# Group 3: Comets and Debris (connected to solar system)
print('\n--- Comets & Debris Group ---')
halley_comet = create_note("Halley's Comet", 'Periodic comet visible from Earth every 75-76 years', 'comet')
oort_cloud = create_note('Oort Cloud', 'Hypothetical spherical cloud of icy bodies', 'debris')
kuiper_belt = create_note('Kuiper Belt', 'Region beyond Neptune with icy bodies and dwarf planets', 'debris')

if all([halley_comet, oort_cloud, kuiper_belt, solar_system]):
    create_link(solar_system, kuiper_belt, 0.7, 'reference')
    create_link(kuiper_belt, oort_cloud, 0.8, 'related')
    create_link(oort_cloud, halley_comet, 0.9, 'dependency')

# Group 4: Unrelated Topics (isolated nodes)
print('\n--- Unrelated Topics (Isolated) ---')
cooking = create_note('Italian Cuisine', 'Traditional Italian cooking methods and recipes', 'planet')
programming = create_note('Python Programming', 'High-level programming language for general-purpose', 'star')
music = create_note('Jazz History', 'Origins and evolution of jazz music genre', 'comet')

# Group 5: More asteroids and debris (some connected)
print('\n--- Asteroids & Space Debris ---')
ceres = create_note('Ceres', 'Largest object in asteroid belt, dwarf planet', 'asteroid')
vesta = create_note('Vesta', 'Second most massive object in asteroid belt', 'asteroid')
space_junk = create_note('Space Debris', 'Artificial objects in orbit around Earth', 'debris')

if all([ceres, vesta, space_junk, asteroid_belt, earth]):
    create_link(asteroid_belt, ceres, 0.9, 'reference')
    create_link(asteroid_belt, vesta, 0.9, 'reference')
    create_link(earth, space_junk, 0.5, 'dependency')

# Cross-theme link (weak connection between groups)
if black_hole and programming:
    create_link(black_hole, programming, 0.2, 'custom')

print('\n=== Test Data Created ===')
print('Open http://localhost:5173 to view the graph')
print('Features to check:')
print('- 2D/3D graph toggle works')
print('- Full graph / Local view toggle works')
print('- Different celestial body types render correctly')
print('- Links have different colors/styles by type')
print('- Isolated nodes (cooking, music) appear correctly')
