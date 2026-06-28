// Test script to check API link data structure
// This simulates what the API might return

const apiResponse = {
  nodes: [
    { id: '1', title: 'Node 1', type: 'star' },
    { id: '2', title: 'Node 2', type: 'planet' },
    { id: '3', title: 'Node 3', type: 'comet' }
  ],
  links: [
    // Different possible field names from backend
    { source_note_id: '1', target_note_id: '2', weight: 0.5, link_type: 'reference' },
    { source: '2', target: '3', weight: 0.3, link_type: 'related' }
  ]
};

// Transformation logic from +page.svelte
const transformedLinks = apiResponse.links.map(l => ({
  source: l.source_note_id || l.source,
  target: l.target_note_id || l.target,
  weight: l.weight,
  link_type: l.link_type
}));

console.log('API Response:', JSON.stringify(apiResponse, null, 2));
console.log('Transformed Links:', JSON.stringify(transformedLinks, null, 2));

// Check if links are valid
const nodeIds = new Set(apiResponse.nodes.map(n => n.id));
console.log('Node IDs:', Array.from(nodeIds));

transformedLinks.forEach((link, i) => {
  const valid = nodeIds.has(link.source) && nodeIds.has(link.target);
  console.log(`Link ${i}: ${link.source} -> ${link.target} (${valid ? 'VALID' : 'INVALID'})`);
});