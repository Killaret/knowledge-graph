/**
 * Wire each CelestialBody instance to its Canvas renderer.
 * Kept in the renderer module so the domain object stays free of Canvas imports.
 */
import { CelestialBody } from "$entities";
import { getAnomalyParams } from "$shared/lib/graph/renderer/anomalies/helpers";
import { drawRealityRift } from "$shared/lib/graph/renderer/anomalies/reality-rift";
import { drawChromaticMaw } from "$shared/lib/graph/renderer/anomalies/chromatic-maw";
import { drawVoidWhisper } from "$shared/lib/graph/renderer/anomalies/void-whisper";
import { drawCosmicAbomination } from "$shared/lib/graph/renderer/anomalies/cosmic-abomination";
import {
  drawStar,
  drawPlanet,
  drawComet,
  drawGalaxy,
  drawNebula,
  drawAsteroid,
  drawDebris,
  drawDust,
  drawBlackhole,
  drawTechnicalNode,
  drawMoon,
  drawUnknown,
} from "./node-renderers";

export function registerCelestialBodyDrawers(): void {
  CelestialBody.STAR.drawFunction = (ctx, c) => {
    drawStar(ctx, c.x, c.y, c.r, c.angle, c.variation, c.nodeId, c.nodeCount, c.time);
  };

  CelestialBody.PLANET.drawFunction = (ctx, c) => {
    drawPlanet(ctx, c.x, c.y, c.r, c.angle, c.variation, c.nodeId, c.nodeCount, c.time);
  };

  CelestialBody.SATELLITE.drawFunction = (ctx, c) => {
    drawPlanet(ctx, c.x, c.y, c.r, c.angle, c.variation, c.nodeId, c.nodeCount, c.time);
  };

  CelestialBody.COMET.drawFunction = (ctx, c) => {
    drawComet(ctx, c.x, c.y, c.r, c.angle, c.variation, c.nodeId, c.nodeCount, c.time);
  };

  CelestialBody.GALAXY.drawFunction = (ctx, c) => {
    drawGalaxy(ctx, c.x, c.y, c.r, c.angle, c.variation, c.nodeId, c.nodeCount, c.time);
  };

  CelestialBody.NEBULA.drawFunction = (ctx, c) => {
    drawNebula(ctx, c.x, c.y, c.r, c.angle, c.variation);
  };

  CelestialBody.ASTEROID.drawFunction = (ctx, c) => {
    drawAsteroid(
      ctx,
      c.x,
      c.y,
      c.r,
      c.angle,
      c.variation,
      c.disableVariation || c.focusMode,
      c.nodeId,
      c.nodeCount,
      c.time
    );
  };

  CelestialBody.DEBRIS.drawFunction = (ctx, c) => {
    drawDebris(
      ctx,
      c.x,
      c.y,
      c.r,
      c.angle,
      c.disableVariation || c.focusMode,
      c.nodeId,
      c.variation
    );
  };

  CelestialBody.DUST.drawFunction = (ctx, c) => {
    drawDust(ctx, c.x, c.y, c.r, c.angle, c.disableVariation || c.focusMode, c.nodeId, c.variation);
  };

  CelestialBody.BLACKHOLE.drawFunction = (ctx, c) => {
    drawBlackhole(ctx, c.x, c.y, c.r, c.angle, c.nodeId, c.nodeCount, c.time, c.variation);
  };

  CelestialBody.TECHNICAL.drawFunction = (ctx, c) => {
    drawTechnicalNode(ctx, c.x, c.y, c.r, c.time, c.variation);
  };

  CelestialBody.MOON.drawFunction = (ctx, c) => {
    drawMoon(ctx, c.x, c.y, c.r, c.angle, c.variation);
  };

  CelestialBody.UNKNOWN.drawFunction = (ctx, c) => {
    drawUnknown(ctx, c.x, c.y, c.r, c.angle, c.nodeId);
  };

  CelestialBody.REALITY_RIFT.drawFunction = (ctx, c) => {
    drawRealityRift(ctx, c.x, c.y, c.r, getAnomalyParams(c.nodeId));
  };

  CelestialBody.CHROMATIC_MAW.drawFunction = (ctx, c) => {
    drawChromaticMaw(ctx, c.x, c.y, c.r, getAnomalyParams(c.nodeId));
  };

  CelestialBody.VOID_WHISPER.drawFunction = (ctx, c) => {
    drawVoidWhisper(ctx, c.x, c.y, c.r, getAnomalyParams(c.nodeId));
  };

  CelestialBody.COSMIC_ABOMINATION.drawFunction = (ctx, c) => {
    drawCosmicAbomination(ctx, c.x, c.y, c.r, getAnomalyParams(c.nodeId));
  };
}
