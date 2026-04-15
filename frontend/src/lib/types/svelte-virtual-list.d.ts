declare module 'svelte-virtual-list' {
  import { SvelteComponent } from 'svelte';
  
  interface VirtualListProps {
    items: any[];
    itemHeight: number;
  }
  
  export default class VirtualList extends SvelteComponent<VirtualListProps> {
    $$prop_def: VirtualListProps;
  }
}
