import requests
import json

BASE_URL = 'http://localhost:8080'

def clear_all_data():
    """Очистка всех связей и заметок"""
    print('🧹 Clearing all data...')
    
    # Получаем все заметки
    resp = requests.get(f'{BASE_URL}/notes?limit=10000')
    if resp.status_code == 200:
        data = resp.json()
        notes = data.get('notes', [])
        print(f'   Found {len(notes)} notes to delete')
        
        # Удаляем все связи для каждой заметки
        for note in notes:
            note_id = note['id']
            links_resp = requests.get(f'{BASE_URL}/notes/{note_id}/links')
            if links_resp.status_code == 200:
                links_data = links_resp.json()
                for link in links_data.get('outgoing', []):
                    requests.delete(f'{BASE_URL}/links/{link["id"]}')
                for link in links_data.get('incoming', []):
                    requests.delete(f'{BASE_URL}/links/{link["id"]}')
        
        # Удаляем все заметки
        deleted = 0
        for note in notes:
            note_id = note['id']
            del_resp = requests.delete(f'{BASE_URL}/notes/{note_id}')
            if del_resp.status_code in [200, 204]:
                deleted += 1
        
        print(f'   ✅ Deleted {deleted} notes')
    
    print('🗑️  Database cleared!\n')


def create_test_data():
    """Создание тестовых данных - минимальный набор для тестов и визуализации"""
    print('🌱 Creating test data...\n')
    
    # Создаем 10 заметок разных типов
    notes_data = [
        {'title': 'Solar System', 'content': 'Our home planetary system with the Sun at center', 'type': 'star'},
        {'title': 'Mercury', 'content': 'First planet from the Sun, smallest and fastest', 'type': 'planet'},
        {'title': 'Venus', 'content': 'Second planet, hottest due to greenhouse effect', 'type': 'planet'},
        {'title': 'Earth', 'content': 'Third planet, only known to support life', 'type': 'planet'},
        {'title': 'Mars', 'content': 'Fourth planet, red due to iron oxide', 'type': 'planet'},
        {'title': 'Jupiter', 'content': 'Fifth planet, largest gas giant with Great Red Spot', 'type': 'planet'},
        {'title': 'Asteroid Belt', 'content': 'Region between Mars and Jupiter with rocky bodies', 'type': 'asteroid'},
        {'title': 'Kuiper Belt', 'content': 'Region beyond Neptune with icy bodies and dwarf planets', 'type': 'debris'},
        {'title': 'Milky Way', 'content': 'Our home galaxy, a barred spiral galaxy', 'type': 'galaxy'},
        {'title': 'Andromeda', 'content': 'Nearest major galaxy to Milky Way, approaching us', 'type': 'galaxy'},
    ]
    
    note_ids = []
    for data in notes_data:
        try:
            resp = requests.post(f'{BASE_URL}/notes', json=data, timeout=5)
            if resp.status_code in [200, 201]:
                note_ids.append(resp.json()['id'])
                print(f'   ✅ {data["title"]} ({data["type"]})')
            else:
                print(f'   ❌ Error creating {data["title"]}: {resp.status_code}')
        except Exception as e:
            print(f'   ❌ Error: {e}')
    
    print(f'\n📊 Created {len(note_ids)} notes\n')
    
    # Создаем связи разных типов
    links_config = [
        # reference links (blue, solid) - orbital relationships
        {'source': 0, 'target': 1, 'link_type': 'reference', 'weight': 0.9},  # Sun -> Mercury
        {'source': 0, 'target': 2, 'link_type': 'reference', 'weight': 0.9},  # Sun -> Venus
        {'source': 0, 'target': 3, 'link_type': 'reference', 'weight': 0.9},  # Sun -> Earth
        
        # dependency links (orange, dashed) - gravitational influence
        {'source': 0, 'target': 4, 'link_type': 'dependency', 'weight': 0.8},  # Sun -> Mars
        {'source': 0, 'target': 5, 'link_type': 'dependency', 'weight': 0.8},  # Sun -> Jupiter
        
        # related links (gray, solid/dashed by weight) - nearby regions
        {'source': 4, 'target': 6, 'link_type': 'related', 'weight': 0.6},   # Mars -> Asteroid Belt
        {'source': 5, 'target': 6, 'link_type': 'related', 'weight': 0.6},   # Jupiter -> Asteroid Belt
        {'source': 5, 'target': 7, 'link_type': 'related', 'weight': 0.5},   # Jupiter -> Kuiper Belt
        
        # weak related links (will be dashed due to weight < 0.3)
        {'source': 8, 'target': 9, 'link_type': 'related', 'weight': 0.25},  # Milky Way -> Andromeda (weak)
        {'source': 3, 'target': 7, 'link_type': 'related', 'weight': 0.2},   # Earth -> Kuiper Belt (weak)
    ]
    
    print('🔗 Creating links...')
    links_created = 0
    for link in links_config:
        source_idx = link['source']
        target_idx = link['target']
        
        if source_idx < len(note_ids) and target_idx < len(note_ids):
            link_data = {
                'source_note_id': note_ids[source_idx],
                'target_note_id': note_ids[target_idx],
                'link_type': link['link_type'],
                'weight': link['weight']
            }
            try:
                resp = requests.post(f'{BASE_URL}/links', json=link_data, timeout=5)
                if resp.status_code in [200, 201]:
                    links_created += 1
                    link_type = link['link_type']
                    weight = link['weight']
                    print(f'   ✅ {notes_data[source_idx]["title"]} -> {notes_data[target_idx]["title"]} ({link_type}, w={weight})')
                else:
                    print(f'   ❌ Error creating link: {resp.status_code} - {resp.text[:100]}')
            except Exception as e:
                print(f'   ❌ Error: {e}')
    
    print(f'\n📊 Created {links_created} links\n')
    
    # Summary
    print('='*50)
    print('📋 SUMMARY:')
    print(f'   📝 Notes: {len(note_ids)}')
    print(f'      • Stars: 1')
    print(f'      • Planets: 5')
    print(f'      • Asteroid belts: 1')
    print(f'      • Debris fields: 1')
    print(f'      • Galaxies: 2')
    print(f'   🔗 Links: {links_created}')
    print(f'      • Reference (blue, solid): 3')
    print(f'      • Dependency (orange, dashed): 2')
    print(f'      • Related (gray, solid): 4')
    print(f'      • Weak related (gray, dashed): 2')
    print('='*50)
    
    return len(note_ids), links_created


if __name__ == '__main__':
    print('\n' + '='*50)
    print('🚀 KNOWLEDGE GRAPH - DATA RESET')
    print('='*50 + '\n')
    
    # Очистка
    clear_all_data()
    
    # Создание новых данных
    notes_count, links_count = create_test_data()
    
    print('\n' + '='*50)
    print(f'✅ Done! Database ready with {notes_count} notes and {links_count} links.')
    print('='*50 + '\n')
