import requests
import json
import random

BASE_URL = 'http://localhost:8080'

# Types of celestial bodies
types = ['star', 'planet', 'comet', 'galaxy', None]

# Create 50 notes
note_ids = []
print('Creating 50 notes...')
for i in range(50):
    note_type = random.choice(types)
    data = {
        'title': f'Test Note {i+1} - {random.choice(["Alpha", "Beta", "Gamma", "Delta", "Epsilon"])} {random.choice(["System", "Nebula", "Cluster", "Field", "Stream"])}',
        'content': f'This is test content for note {i+1}. ' + random.choice([
            'Contains astronomical data and observations.',
            'Linked to stellar evolution research.',
            'References multiple star catalogs.',
            'Part of galactic survey project.',
            'Contains spectral analysis data.'
        ]),
        'type': note_type if note_type else 'star'
    }
    try:
        resp = requests.post(f'{BASE_URL}/notes', json=data, timeout=5)
        if resp.status_code == 200 or resp.status_code == 201:
            note_ids.append(resp.json()['id'])
            if (i + 1) % 10 == 0:
                print(f'  Created {i+1} notes...')
        else:
            print(f'  Error creating note {i+1}: {resp.status_code}')
    except Exception as e:
        print(f'  Error: {e}')

print(f'\nCreated {len(note_ids)} notes')

# Create different types of links
print('\nCreating links...')
links_created = 0

# 1. Strong direct links (first 15 notes form a chain)
for i in range(min(14, len(note_ids)-1)):
    try:
        data = {
            'source_note_id': note_ids[i],
            'target_note_id': note_ids[i+1],
            'link_type': 'dependency',
            'weight': round(random.uniform(0.8, 1.0), 2)
        }
        resp = requests.post(f'{BASE_URL}/links', json=data, timeout=5)
        if resp.status_code in [200, 201]:
            links_created += 1
    except Exception as e:
        pass

print(f'  Created {links_created} strong chain links (notes 1-15)')

# 2. Medium semantic links (notes 15-30 connect randomly)
for i in range(15, min(30, len(note_ids))):
    for j in range(random.randint(1, 3)):
        target = random.randint(15, min(35, len(note_ids)-1))
        if target != i and target < len(note_ids):
            try:
                data = {
                    'source_note_id': note_ids[i],
                    'target_note_id': note_ids[target],
                    'link_type': 'reference',
                    'weight': round(random.uniform(0.4, 0.7), 2)
                }
                resp = requests.post(f'{BASE_URL}/links', json=data, timeout=5)
                if resp.status_code in [200, 201]:
                    links_created += 1
            except:
                pass

print(f'  Created medium semantic links (notes 15-30)')

# 3. Weak links (notes 30-40 with few connections)
weak_links = 0
for i in range(30, min(40, len(note_ids))):
    if random.random() > 0.5:
        target = random.randint(0, len(note_ids)-1)
        if target != i:
            try:
                data = {
                    'source_note_id': note_ids[i],
                    'target_note_id': note_ids[target],
                    'link_type': 'related',
                    'weight': round(random.uniform(0.1, 0.3), 2)
                }
                resp = requests.post(f'{BASE_URL}/links', json=data, timeout=5)
                if resp.status_code in [200, 201]:
                    weak_links += 1
            except:
                pass

print(f'  Created {weak_links} weak links (notes 30-40)')

# 4. Notes 40-50 have NO links (isolated)
print(f'  Notes 41-50 are isolated (no links)')

print(f'\nTotal links created: {links_created + weak_links}')
print('\nTest data structure:')
print(f'  - Notes 1-15: Strong connected chain')
print(f'  - Notes 15-30: Medium semantic connections')
print(f'  - Notes 30-40: Weak connections')
print(f'  - Notes 40-50: Isolated (no links)')
