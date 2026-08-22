const baseUrl = process.argv[2] || "http://localhost:5173";
const urls = [
  `${baseUrl}/_app/immutable/nodes/0.BlRZLei0.js`,
  `${baseUrl}/_app/immutable/chunks/DGwQnvsv.js`,
  `${baseUrl}/_app/immutable/chunks/2o-GrsBY.js`,
  `${baseUrl}/_app/immutable/chunks/CjFILxn0.js`,
  `${baseUrl}/_app/immutable/chunks/B2Gpw_fT.js`,
  `${baseUrl}/_app/immutable/chunks/B5jsVAZ1.js`,
  `${baseUrl}/_app/immutable/chunks/C5d5ZEH2.js`,
  `${baseUrl}/_app/immutable/chunks/CZsKW0oP.js`,
  `${baseUrl}/_app/immutable/chunks/CXQSA9XB.js`,
  `${baseUrl}/_app/immutable/chunks/Cuks-R1Y.js`,
  `${baseUrl}/_app/immutable/nodes/1.B6-Jdu5Q.js`,
  `${baseUrl}/_app/immutable/nodes/19.CpBxjWD8.js`,
  `${baseUrl}/_app/immutable/chunks/BFuupt00.js`,
  `${baseUrl}/_app/immutable/chunks/SeHWovpS.js`,
];

(async () => {
  for (const url of urls) {
    try {
      const res = await fetch(url);
      const text = await res.text();
      const markers = ["d6aa5d", "drawPlanet", "console.log", "#60a5fa"];
      const foundMarkers = markers.filter((m) => text.includes(m));
      if (foundMarkers.length > 0) {
        console.log("FOUND in", url, "markers=", foundMarkers);
        for (const m of foundMarkers) {
          let idx = text.indexOf(m);
          while (idx !== -1) {
            const start = Math.max(0, idx - 80);
            const end = Math.min(text.length, idx + m.length + 80);
            console.log(`  MATCH ${m} at ${idx}:`, text.slice(start, end).replace(/\n/g, " "));
            idx = text.indexOf(m, idx + 1);
          }
        }
      }
    } catch (e) {
      console.error("ERROR", url, e.message);
    }
  }
})();
