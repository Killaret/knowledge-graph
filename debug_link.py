import requests

# Get recent notes
resp = requests.get('http://localhost:8080/notes?limit=5')
data = resp.json()
print(f"Total notes: {data['total']}")
if data['notes']:
    id1 = data['notes'][0]['id']
    id2 = data['notes'][1]['id']
    print(f"Note 1: {id1}")
    print(f"Note 2: {id2}")
    
    # Try to create link
    link_data = {
        'source_note_id': id1,
        'target_note_id': id2,
        'link_type': 'strong',
        'weight': 0.8
    }
    print(f"Sending: {link_data}")
    resp2 = requests.post('http://localhost:8080/links', json=link_data)
    print(f"Status: {resp2.status_code}")
    print(f"Response: {resp2.text}")
