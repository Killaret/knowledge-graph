import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from '@sveltejs/kit';
import { createNote } from '$lib/api/notes';

export const actions: Actions = {
  default: async ({ request }) => {
    try {
      const data = await request.formData();
      const title = data.get('title');
      const content = data.get('content');
      
      if (!title) {
        return fail(400, { message: 'Title is required' });
      }
      
      const note = await createNote({ 
        title: title.toString(), 
        content: content?.toString() || '', 
        metadata: {} 
      });
      
      throw redirect(302, `/notes/${note.id}`);
    } catch (e) {
      console.error('Server action create error:', e);
      return fail(500, { message: 'Failed to create note' });
    }
  }
};
