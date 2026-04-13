import requests

BASE_URL = 'http://localhost:8080'

# Clear all notes
resp = requests.get(f'{BASE_URL}/notes?limit=10000')
if resp.status_code == 200:
    data = resp.json()
    notes = data.get('notes', [])
    print(f'Found {len(notes)} notes to delete')
    
    for note in notes:
        note_id = note['id']
        # Delete links
        links_resp = requests.get(f'{BASE_URL}/notes/{note_id}/links')
        if links_resp.status_code == 200:
            links_data = links_resp.json()
            for link in links_data.get('outgoing', []):
                link_id = link.get('id')
                if link_id:
                    requests.delete(f'{BASE_URL}/links/{link_id}')
            for link in links_data.get('incoming', []):
                link_id = link.get('id')
                if link_id:
                    requests.delete(f'{BASE_URL}/links/{link_id}')
        # Delete note
        requests.delete(f'{BASE_URL}/notes/{note_id}')
    
    print(f'Deleted {len(notes)} notes')

print('Database cleared!')
