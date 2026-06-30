# Manual Testing Checklist - Complete Version (FULL VERSION)

## 🎯 Overview

This checklist is designed for comprehensive manual testing of the Knowledge Graph application, including background data preloading, 3D graph with fog effect, and authentication system.

---

## 📋 Preparing for Testing

### 🔧 Environment
- [ ] Clear browser cache (Ctrl+Shift+Delete)
- [ ] Open Developer Tools (F12)
- [ ] Enable Network tab for request monitoring
- [ ] Enable Console tab for log tracking
- [ ] Ensure application is running in development mode

### 📝 Test Data
- [ ] Prepare test credentials:
  - Email: `test@example.com`
  - Password: `testpassword`
- [ ] Verify test data exists in knowledge base
- [ ] Ensure connections exist between notes

---

## 🚀 1. PreloadService Testing

### 1.1 Cold Start
**HOW TO TEST:**
- [ ] **Step 1:** Open browser in incognito mode (Ctrl+Shift+N)
- [ ] **Step 2:** Enter application URL and press Enter
- [ ] **Step 3:** Open Developer Tools (F12) → Console tab
- [ ] **Step 4:** Find message `[PreloadService] Starting preload`
- [ ] **Step 5:** Go to Network tab → filter by `api/`
- [ ] **Step 6:** Verify requests are visible:
  - `GET /api/graph/full` 
  - `GET /api/users/achievements`
- [ ] **Step 7:** Verify requests are executed in parallel
- [ ] **Step 8:** Verify cache TTL (5 min for graph, 10 min for achievements)

**EXPECTED RESULT:** Requests are executed, data is cached, console messages appear

### 1.2 Preloaded Data (Cached Data)
**HOW TO TEST:**
- [ ] **Step 1:** Log in to application with test data
- [ ] **Step 2:** Wait for full graph loading
- [ ] **Step 3:** Click "Logout" button
- [ ] **Step 4:** Refresh page (F5)
- [ ] **Step 5:** Check console: should show `[PreloadService] Using cached data`
- [ ] **Step 6:** Check Network tab - no new API requests should appear
- [ ] **Step 7:** Verify data displays instantly (<1 second)

**EXPECTED RESULT:** Instant loading from cache, no network requests

### 1.3 Cache Clearing
**HOW TO TEST:**
- [ ] **Step 1:** While in application, click "Logout"
- [ ] **Step 2:** Check console: should show `[PreloadService] Cache cleared`
- [ ] **Step 3:** Check localStorage → cache keys should be absent
- [ ] **Step 4:** Refresh page (F5)
- [ ] **Step 5:** Check Network tab - API requests should appear again

**EXPECTED RESULT:** Cache cleared on logout, requests executed again

---

## 🌫 2. Fog Veil Effect Testing

### 2.1 Initial Fog
**HOW TO TEST:**
- [ ] **Step 1:** Logout and clear cache (Ctrl+Shift+Delete)
- [ ] **Step 2:** Go to application main page
- [ ] **Step 3:** Open Developer Tools → Console tab
- [ ] **Step 4:** Find message `[Graph3D] Starting simulation with preloaded data: false, initial fog density: 0.08`
- [ ] **Step 5:** Visually evaluate fog density - should be dense, hiding distant nodes
- [ ] **Step 6:** Check fog color - should match background (#050510)

**EXPECTED RESULT:** Dense fog appears immediately on graph loading

### 2.2 Progressive Dissipation
**HOW TO TEST:**
- [ ] **Step 1:** Observe 3D graph after simulation starts
- [ ] **Step 2:** Monitor console progress messages: `[Graph3D] Fog progress: X.X%, density: X.XXXX`
- [ ] **Step 3:** Note density change moments (every 5 ticks)
- [ ] **Step 4:** Visually evaluate gradual fog dissipation
- [ ] **Step 5:** Verify nodes become visible as they are positioned

**EXPECTED RESULT:** Smooth density decrease from 0.08 to ~0.005 as progress advances

### 2.3 Final Animation
**HOW TO TEST:**
- [ ] **Step 1:** Wait for simulation completion (when all nodes are positioned)
- [ ] **Step 2:** Observe smooth dissipation animation (~800ms duration)
- [ ] **Step 3:** Check final density - should be ~0.005 (almost transparent)
- [ ] **Step 4:** Verify all nodes are clearly visible without fog
- [ ] **Step 5:** Verify animation uses ease-out cubic effect

**EXPECTED RESULT:** Smooth transition to almost transparent fog

### 2.4 Integration with PreloadService
**HOW TO TEST:**
- [ ] **Step 1:** Log in and wait for full loading
- [ ] **Step 2:** Logout and login again immediately (data should be cached)
- [ ] **Step 3:** Check console: `[Graph3D] Starting simulation with preloaded data: true, initial fog density: 0.04`
- [ ] **Step 4:** Visually evaluate initial density - should be less than cold start
- [ ] **Step 5:** Verify effect still works correctly

**EXPECTED RESULT:** Less dense initial fog for cached data

### 2.5 Edge Cases
**HOW TO TEST:**
- [ ] **Empty graph:**
  - **Step 1:** Create/load graph with 0 nodes
  - **Step 2:** Verify fog immediately sets to 0.005
  - **Step 3:** Ensure no dissipation animation
- [ ] **Single node:**
  - **Step 1:** Load graph with 1 node
  - **Step 2:** Verify immediate density set to 0.005
- [ ] **Large graph (>100 nodes):**
  - **Step 1:** Load graph with many nodes
  - **Step 2:** Check animation performance
  - **Step 3:** Ensure fog dissipates smoothly
- [ ] **Graph without connections:**
  - **Step 1:** Load nodes without connections
  - **Step 2:** Verify effect works correctly

**EXPECTED RESULT:** Correct operation in all edge cases

---

## 🔐 3. Authentication System Testing

### 3.1 Login
**HOW TO TEST:**
- [ ] **Step 1:** Open browser → enter application URL → navigate to `/auth/login`
- [ ] **Step 2:** Enter test credentials:
  - Email: `test@example.com`
  - Password: `testpassword`
- [ ] **Step 3:** Click "Login" button
- [ ] **Step 4:** Verify redirect to main page (`/`)
- [ ] **Step 5:** Open Developer Tools → Console → verify no errors
- [ ] **Step 6:** Check localStorage → tokens should appear:
  - `access_token`
  - `refresh_token`
  - `user` data
- [ ] **Step 7:** Verify PreloadService cache is NOT cleared on login

**EXPECTED RESULT:** Successful login, redirect, tokens saved

### 3.2 Logout
**HOW TO TEST:**
- [ ] **Step 1:** While in application, find "Logout" button
- [ ] **Step 2:** Click logout button
- [ ] **Step 3:** Verify redirect to login page (`/auth/login`)
- [ ] **Step 4:** Check console: should show `[PreloadService] Cache cleared`
- [ ] **Step 5:** Check localStorage → tokens should be removed
- [ ] **Step 6:** Verify PreloadService cache is cleared:
  - `graph_cache` absent
  - `achievements_cache` absent

**EXPECTED RESULT:** Correct logout, all data cleared

### 3.3 Token Refresh
**HOW TO TEST:**
- [ ] **Step 1:** Log in with valid credentials
- [ ] **Step 2:** Open Developer Tools → Application → Local Storage
- [ ] **Step 3:** Find `access_token` and change it to expired (or delete)
- [ ] **Step 4:** Perform any action requiring authorization
- [ ] **Step 5:** Observe Network tab → token refresh request should appear
- [ ] **Step 6:** Verify user stays in system (not kicked to login)
- [ ] **Step 7:** Verify PreloadService cache is preserved after refresh

**EXPECTED RESULT:** Automatic token refresh without session break

---

## 📊 4. 3D Graph Testing

### 4.1 Basic Functionality
**HOW TO TEST:**
- [ ] **Camera rotation:**
  - **Step 1:** Hover cursor over 3D graph
  - **Step 2:** Hold left mouse button and move left/right
  - **Step 3:** Verify camera rotates around graph center
  - **Step 4:** Verify smooth rotation without jerks
- [ ] **Zooming:**
  - **Step 1:** Hover cursor over 3D graph
  - **Step 2:** Use mouse wheel to scroll up
  - **Step 3:** Verify zoom in
  - **Step 4:** Scroll wheel down - verify zoom out
- [ ] **Panning:**
  - **Step 1:** Hold right mouse button
  - **Step 2:** Move mouse left/right/up/down
  - **Step 3:** Verify camera moves in corresponding direction
- [ ] **Auto-rotation:**
  - **Step 1:** Verify graph slowly rotates without interaction
  - **Step 2:** Start manual rotation - auto-rotation should stop
  - **Step 3:** Release mouse - auto-rotation should resume
- [ ] **"Reset View" button:**
  - **Step 1:** Find "Reset View" or camera reset button
  - **Step 2:** Click button
  - **Step 3:** Verify camera returns to optimal position

**EXPECTED RESULT:** All basic camera functions work smoothly

### 4.2 Node Interaction
**HOW TO TEST:**
- [ ] **Single click:**
  - **Step 1:** Hover cursor over any graph node
  - **Step 2:** Click left mouse button
  - **Step 3:** Verify node is highlighted (visual effect)
  - **Step 4:** Verify node information appears
- [ ] **Double click:**
  - **Step 1:** Hover cursor over node
  - **Step 2:** Quickly double-click left mouse button
  - **Step 3:** Verify camera centers on this node
  - **Step 4:** Verify smooth centering animation
- [ ] **Information display:**
  - **Step 1:** Click on node
  - **Step 2:** Verify tooltip/panel with information appears
  - **Step 3:** Verify title and node type are displayed
- [ ] **Navigation by connections:**
  - **Step 1:** Find connected node
  - **Step 2:** Verify connection line visibility
  - **Step 3:** Verify ability to navigate to connected node

**EXPECTED RESULT:** Interactive elements work correctly

### 4.3 Performance
**HOW TO TEST:**
- [ ] **Animation smoothness:**
  - **Step 1:** Open Developer Tools → Rendering → FPS counter
  - **Step 2:** Rotate camera and observe FPS
  - **Step 3:** Ensure FPS doesn't drop below 30
  - **Step 4:** Verify no jerks or stuttering
- [ ] **Large graph:**
  - **Step 1:** Load graph with >100 nodes
  - **Step 2:** Check initial loading time
  - **Step 3:** Check performance during interaction
  - **Step 4:** Ensure browser doesn't hang
- [ ] **Memory leaks:**
  - **Step 1:** Open Developer Tools → Memory
  - **Step 2:** Take memory snapshot before graph loading
  - **Step 3:** Interact with graph for 2-3 minutes
  - **Step 4:** Take new snapshot and compare memory growth
- [ ] **Responsiveness:**
  - **Step 1:** Change browser window size
  - **Step 2:** Verify 3D graph adapts to new size
  - **Step 3:** Verify correct proportions and scale

**EXPECTED RESULT:** High performance without memory leaks

---

## 🌐 5. Responsiveness and UI Testing

### 5.1 Mobile Devices
**HOW TO TEST:**
- [ ] **Mobile mode:**
  - **Step 1:** Open Chrome DevTools (F12)
  - **Step 2:** Click device icon (Toggle device toolbar)
  - **Step 3:** Select iPhone 14 or Android
  - **Step 4:** Verify correct display on small screen
- [ ] **Touch gestures:**
  - **Step 1:** Try rotating camera with swipe
  - **Step 2:** Verify pinch-to-zoom scaling
  - **Step 3:** Verify panning with swipe
- [ ] **UI adaptation:**
  - **Step 1:** Check button sizes on mobile screen
  - **Step 2:** Check menu and panel visibility
  - **Step 3:** Ensure elements don't go off screen

**EXPECTED RESULT:** Correct operation on mobile devices

### 5.2 Different Resolutions
**HOW TO TEST:**
- [ ] **Desktop (1920x1080):**
  - **Step 1:** Open DevTools → Responsive Design Mode
  - **Step 2:** Select resolution 1920x1080
  - **Step 3:** Verify correct display
- [ ] **Laptop (1366x768):**
  - **Step 1:** Select resolution 1366x768
  - **Step 2:** Check UI element adaptation
  - **Step 3:** Ensure no horizontal scrolling
- [ ] **Tablet (768x1024):**
  - **Step 1:** Select resolution 768x1024
  - **Step 2:** Check interface layout
  - **Step 3:** Check control element availability
- [ ] **Mobile (375x667):**
  - **Step 1:** Select resolution 375x667
  - **Step 2:** Check vertical layout
  - **Step 3:** Ensure touch elements are correct

**EXPECTED RESULT:** Responsiveness on all resolutions

### 5.3 Dark/Light Theme
**HOW TO TEST:**
- [ ] **Dark theme:**
  - **Step 1:** Ensure application is in dark theme
  - **Step 2:** Check UI element contrast
  - **Step 3:** Ensure 3D graph is clearly visible on dark background
  - **Step 4:** Check text and label readability
- [ ] **Light theme (if available):**
  - **Step 1:** Switch to light theme
  - **Step 2:** Verify correct display
  - **Step 3:** Ensure no contrast issues
- [ ] **Theme switching:**
  - **Step 1:** Find theme switcher
  - **Step 2:** Switch between themes
  - **Step 3:** Verify smooth transition

**EXPECTED RESULT:** Correct display in both themes

---

## 🔧 6. Error and Edge Case Testing

### 6.1 Network Errors
**HOW TO TEST:**
- [ ] **Internet disconnection:**
  - **Step 1:** Disconnect Wi-Fi/network
  - **Step 2:** Open application
  - **Step 3:** Verify network error display
  - **Step 4:** Ensure no infinite loading
- [ ] **Connection recovery:**
  - **Step 1:** Restore internet connection
  - **Step 2:** Verify automatic loading recovery
  - **Step 3:** Ensure application continues working
- [ ] **Slow connection:**
  - **Step 1:** Open Chrome DevTools → Network → Slow 3G
  - **Step 2:** Reload application
  - **Step 3:** Verify slow loading indicator display

**EXPECTED RESULT:** Correct network error handling

### 6.2 Incorrect Data
**HOW TO TEST:**
- [ ] **Invalid login:**
  - **Step 1:** Go to login page
  - **Step 2:** Enter invalid email: `wrong@example.com`
  - **Step 3:** Enter invalid password
  - **Step 4:** Click "Login"
  - **Step 5:** Verify authentication error display
- [ ] **Non-existent page:**
  - **Step 1:** Enter URL of non-existent page
  - **Step 2:** Verify 404 error display
  - **Step 3:** Ensure back to home button exists
- [ ] **Empty data:**
  - **Step 1:** Try login with empty fields
  - **Step 2:** Check form validation
  - **Step 3:** Ensure hints are displayed

**EXPECTED RESULT:** Correct handling of incorrect data

### 6.3 Browser Compatibility
**HOW TO TEST:**
- [ ] **Chrome:**
  - **Step 1:** Open application in Chrome (latest version)
  - **Step 2:** Check all main functions
  - **Step 3:** Ensure no console errors
- [ ] **Firefox:**
  - **Step 1:** Open application in Firefox (latest version)
  - **Step 2:** Check 3D graph and UI
  - **Step 3:** Check performance
- [ ] **Safari (if available):**
  - **Step 1:** Open application in Safari
  - **Step 2:** Verify correct display
  - **Step 3:** Check touch gestures
- [ ] **Edge:**
  - **Step 1:** Open application in Edge (latest version)
  - **Step 2:** Check all functions
  - **Step 3:** Compare with Chrome

**EXPECTED RESULT:** Cross-browser compatibility

---

## 📝 7. Logs and Monitoring Testing

### 7.1 Console Logs
**HOW TO TEST:**
- [ ] **Error checking:**
  - **Step 1:** Open Developer Tools → Console
  - **Step 2:** Verify no red errors (ERROR)
  - **Step 3:** Verify no yellow warnings (WARNING)
  - **Step 4:** Ensure no critical issues
- [ ] **Informational logs:**
  - **Step 1:** Find PreloadService messages
  - **Step 2:** Find 3D graph messages
  - **Step 3:** Find authentication system messages
  - **Step 4:** Ensure all logs are informative
- [ ] **Fog logs:**
  - **Step 1:** Find `[Graph3D] Starting simulation` messages
  - **Step 2:** Find `[Graph3D] Fog progress` messages
  - **Step 3:** Verify correct density values

**EXPECTED RESULT:** Clean console without errors

### 7.2 Network Requests
**HOW TO TEST:**
- [ ] **Request statuses:**
  - **Step 1:** Open Developer Tools → Network
  - **Step 2:** Perform actions in application
  - **Step 3:** Verify all requests have 200 status
  - **Step 4:** Ensure no 4xx/5xx errors
- [ ] **Request optimization:**
  - **Step 1:** Verify no duplicate requests
  - **Step 2:** Ensure no unnecessary requests
  - **Step 3:** Check cache usage (304 statuses)
- [ ] **Authorization headers:**
  - **Step 1:** Click on any API request
  - **Step 2:** Check Authorization header
  - **Step 3:** Verify correct Bearer token

**EXPECTED RESULT:** Optimal network requests

### 7.3 Performance
**HOW TO TEST:**
- [ ] **Core Web Vitals:**
  - **Step 1:** Open Chrome DevTools → Lighthouse
  - **Step 2:** Run performance analysis
  - **Step 3:** Verify LCP < 2.5s
  - **Step 4:** Verify FID < 100ms
  - **Step 5:** Verify CLS < 0.1
- [ ] **Application size:**
  - **Step 1:** Open Network tab
  - **Step 2:** Refresh page (Ctrl+R)
  - **Step 3:** Check loaded resource size
  - **Step 4:** Ensure size is acceptable
- [ ] **Blocking resources:**
  - **Step 1:** Check Performance tab
  - **Step 2:** Find blocking resources
  - **Step 3:** Ensure critical resources don't block rendering

**EXPECTED RESULT:** High application performance

---

## ✅ 8. Final Check

### 8.1 Complete User Scenario
**HOW TO TEST:**
- [ ] **Cold start → preload → login → 3D graph with fog:**
  - **Step 1:** Open application in incognito mode
  - **Step 2:** Verify data preloading and fog
  - **Step 3:** Log in to system
  - **Step 4:** Verify 3D graph with fog effect works
- [ ] **Logout → cache clear → login → instant loading:**
  - **Step 1:** Logout from application
  - **Step 2:** Verify cache clearing
  - **Step 3:** Login again
  - **Step 4:** Verify instant loading from cache
- [ ] **Graph navigation → node interaction:**
  - **Step 1:** Interact with graph nodes
  - **Step 2:** Verify all camera functions
  - **Step 3:** Verify information display
- [ ] **Page refresh → cache usage:**
  - **Step 1:** Refresh page
  - **Step 2:** Verify cached data usage
  - **Step 3:** Ensure no unnecessary requests

### 8.2 Critical Functions
**HOW TO TEST:**
- [ ] **PreloadService works correctly**
- [ ] **Fog effect appears and dissipates**
- [ ] **Authentication works without errors**
- [ ] **3D graph is functional and performant**
- [ ] **UI is responsive and user-friendly**

---

## 📊 Test Results

### Successfully passed tests: ___/___
### Issues found:
1. 
2. 
3. 

### Recommendations:
1. 
2. 
3. 

### Status: 
- [ ] All tests passed
- [ ] Minor issues exist
- [ ] Critical issues exist

---

## 🔍 Additional Checks

### For Developers
- [ ] Check unit tests work (`npm test`)
- [ ] Check E2E tests work (`npm run test:e2e`)
- [ ] Check code coverage (`npm run test:coverage`)
- [ ] Ensure build passes without errors (`npm run build`)

### For QA
- [ ] Conduct regression testing
- [ ] Check backend compatibility
- [ ] Conduct load testing
- [ ] Check application security

---

## 📞 Feedback Contacts

When issues are found:
- Create issue in project repository
- Specify reproduction steps
- Attach screenshots/video
- Specify browser and version

---

*Checklist created: May 8, 2026*
*Version: 2.0 (Full version with detailed steps)*
*Updated: -*
