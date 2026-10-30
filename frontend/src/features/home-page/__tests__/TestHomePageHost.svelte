<script lang="ts">
  import { createHomePageState, type HomePageState } from "../";

  const homePage = createHomePageState();
  (globalThis as unknown as { __TEST_HOME_PAGE: HomePageState }).__TEST_HOME_PAGE = homePage;
</script>

<div data-testid="home-page-host">loaded</div>
