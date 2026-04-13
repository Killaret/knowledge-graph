import requests
import json
import random

BASE_URL = 'http://localhost:8080'

def clear_all_data():
    """Очистка всех связей и заметок"""
    print('Clearing all data...')
    
    # Получаем все заметки
    resp = requests.get(f'{BASE_URL}/notes?limit=10000')
    if resp.status_code == 200:
        data = resp.json()
        notes = data.get('notes', [])
        print(f'  Found {len(notes)} notes to delete')
        
        # Удаляем все связи для каждой заметки
        for note in notes:
            note_id = note['id']
            # Получаем связи заметки
            links_resp = requests.get(f'{BASE_URL}/notes/{note_id}/links')
            if links_resp.status_code == 200:
                links_data = links_resp.json()
                # Удаляем исходящие связи
                for link in links_data.get('outgoing', []):
                    requests.delete(f'{BASE_URL}/links/{link["id"]}')
                # Удаляем входящие связи
                for link in links_data.get('incoming', []):
                    requests.delete(f'{BASE_URL}/links/{link["id"]}')
        
        # Удаляем все заметки
        deleted = 0
        for note in notes:
            note_id = note['id']
            del_resp = requests.delete(f'{BASE_URL}/notes/{note_id}')
            if del_resp.status_code in [200, 204]:
                deleted += 1
        
        print(f'  Deleted {deleted} notes')
    
    print('Database cleared!')


def create_group_1_direct_links():
    """Группа 1: Заметки с прямыми связями (linked chain)"""
    print('\nGroup 1: Creating linked chain (15 notes with direct connections)')
    
    topics = [
        ('Solar System', 'star', 'Central star of our planetary system'),
        ('Mercury', 'planet', 'First planet from the Sun, smallest in system'),
        ('Venus', 'planet', 'Second planet, hottest due to greenhouse effect'),
        ('Earth', 'planet', 'Third planet, only known to support life'),
        ('Mars', 'planet', 'Fourth planet, red due to iron oxide'),
        ('Jupiter', 'planet', 'Fifth planet, largest gas giant'),
        ('Saturn', 'planet', 'Sixth planet, famous for ring system'),
        ('Uranus', 'planet', 'Seventh planet, ice giant tilted on side'),
        ('Neptune', 'planet', 'Eighth planet, windiest in solar system'),
        ('Pluto', 'comet', 'Dwarf planet in Kuiper belt'),
        ('Asteroid Belt', 'comet', 'Region between Mars and Jupiter'),
        ('Kuiper Belt', 'comet', 'Region beyond Neptune with icy bodies'),
        ('Oort Cloud', 'comet', 'Theoretical cloud of icy objects'),
        ('Milky Way', 'galaxy', 'Our home galaxy containing solar system'),
        ('Andromeda', 'galaxy', 'Nearest spiral galaxy to Milky Way'),
    ]
    
    note_ids = []
    for title, node_type, content in topics:
        data = {
            'title': title,
            'content': content,
            'type': node_type
        }
        resp = requests.post(f'{BASE_URL}/notes', json=data, timeout=5)
        if resp.status_code in [200, 201]:
            note_ids.append(resp.json()['id'])
    
    # Создаем цепочку связей (каждая с каждой следующей)
    links_count = 0
    for i in range(len(note_ids) - 1):
        for j in range(i + 1, min(i + 3, len(note_ids))):  # Связь с 2 следующими
            weight = round(random.uniform(0.8, 1.0), 2)
            link_data = {
                'source_note_id': note_ids[i],
                'target_note_id': note_ids[j],
                'link_type': 'dependency',
                'weight': weight
            }
            resp = requests.post(f'{BASE_URL}/links', json=link_data)
            if resp.status_code in [200, 201]:
                links_count += 1
    
    print(f'  Created {len(note_ids)} notes with {links_count} strong links')
    return note_ids


def create_group_2_semantic():
    """Группа 2: Семантически связанные заметки (похожая тематика)"""
    print('\nGroup 2: Creating semantic cluster (20 notes on stellar physics)')
    
    topics = [
        ('Stellar Evolution', 'star', 'Lifecycle from birth to death of stars'),
        ('Main Sequence', 'star', 'Phase where stars spend most of their life'),
        ('Red Giant', 'star', 'Late stage when star expands and cools'),
        ('White Dwarf', 'star', 'Remnant core after red giant phase'),
        ('Neutron Star', 'star', 'Extremely dense collapsed star core'),
        ('Black Hole', 'star', 'Region where gravity prevents escape of light'),
        ('Supernova', 'comet', 'Explosive death of massive stars'),
        ('Nuclear Fusion', 'star', 'Process powering stars including our Sun'),
        ('Hertzsprung-Russell', 'star', 'Diagram classifying stars by luminosity'),
        ('Stellar Nursery', 'nebula', 'Region where new stars are born'),
        ('Planetary Nebula', 'nebula', 'Shell of gas from dying low-mass stars'),
        ('Supernova Remnant', 'nebula', 'Structure left after supernova explosion'),
        ('Chandrasekhar Limit', 'star', 'Maximum mass of stable white dwarf'),
        ('Tolman-Oppenheimer', 'star', 'Maximum mass of neutron stars'),
        ('Stellar Wind', 'comet', 'Flow of charged particles from stars'),
        ('Accretion Disk', 'star', 'Structure forming around massive objects'),
        ('Gamma Ray Burst', 'comet', 'Most energetic electromagnetic events'),
        ('Magnetar', 'star', 'Neutron star with extremely strong magnetic field'),
        ('Pulsar', 'star', 'Rotating neutron star emitting beams'),
        ('Binary Star', 'star', 'System of two stars orbiting common center'),
    ]
    
    note_ids = []
    for title, node_type, content in topics:
        data = {
            'title': title,
            'content': content,
            'type': node_type if node_type in ['star', 'planet', 'comet', 'galaxy'] else 'star'
        }
        resp = requests.post(f'{BASE_URL}/notes', json=data, timeout=5)
        if resp.status_code in [200, 201]:
            note_ids.append(resp.json()['id'])
    
    # Семантические связи (много связей между похожими темами)
    links_count = 0
    for i in range(len(note_ids)):
        # Каждая заметка связана с 2-4 другими случайными из группы
        for _ in range(random.randint(2, 4)):
            target = random.randint(0, len(note_ids) - 1)
            if target != i:
                weight = round(random.uniform(0.4, 0.7), 2)
                link_data = {
                    'source_note_id': note_ids[i],
                    'target_note_id': note_ids[target],
                    'link_type': 'reference',
                    'weight': weight
                }
                resp = requests.post(f'{BASE_URL}/links', json=link_data)
                if resp.status_code in [200, 201]:
                    links_count += 1
    
    print(f'  Created {len(note_ids)} notes with {links_count} semantic links')
    return note_ids


def create_group_3_isolated():
    """Группа 3: Изолированные заметки на разные темы (без связей)"""
    print('\nGroup 3: Creating isolated notes (15 notes on various topics)')
    
    topics = [
        ('Dark Matter', 'galaxy', 'Hypothetical matter affecting galaxy rotation'),
        ('Dark Energy', 'galaxy', 'Force causing accelerating universe expansion'),
        ('Cosmic Microwave', 'galaxy', 'Radiation remnant from Big Bang'),
        ('Hubble Constant', 'star', 'Rate of universe expansion measurement'),
        ('Gravitational Waves', 'comet', 'Ripples in spacetime from massive events'),
        ('Exoplanets', 'planet', 'Planets orbiting stars outside solar system'),
        (' SETI', 'star', 'Search for extraterrestrial intelligence project'),
        ('Space Telescopes', 'comet', 'Orbital instruments observing universe'),
        ('Rocket Propulsion', 'comet', 'Technology for space travel'),
        ('Orbital Mechanics', 'planet', 'Physics of objects in orbit'),
        ('Astronomical Units', 'star', 'Distance measurements in astronomy'),
        ('Light Year', 'star', 'Distance light travels in one year'),
        ('Parsec', 'star', 'Unit used for interstellar distances'),
        ('Cosmic Rays', 'comet', 'High-energy particles from space'),
        ('Asteroid Mining', 'comet', 'Concept of extracting resources from asteroids'),
    ]
    
    note_ids = []
    for title, node_type, content in topics:
        data = {
            'title': title,
            'content': content,
            'type': node_type
        }
        resp = requests.post(f'{BASE_URL}/notes', json=data, timeout=5)
        if resp.status_code in [200, 201]:
            note_ids.append(resp.json()['id'])
    
    # НЕ создаем связи - эти заметки изолированы
    print(f'  Created {len(note_ids)} isolated notes (no links)')
    return note_ids


if __name__ == '__main__':
    # Очистка
    clear_all_data()
    
    # Создание трех групп
    group1 = create_group_1_direct_links()
    group2 = create_group_2_semantic()
    group3 = create_group_3_isolated()
    
    # Итог
    total_notes = len(group1) + len(group2) + len(group3)
    print(f'\n{"="*50}')
    print(f'SUMMARY:')
    print(f'  Group 1 (Direct links): {len(group1)} notes - Solar System chain')
    print(f'  Group 2 (Semantic):     {len(group2)} notes - Stellar Physics cluster')
    print(f'  Group 3 (Isolated):       {len(group3)} notes - Various topics')
    print(f'  TOTAL: {total_notes} notes created')
    print(f'{"="*50}')
