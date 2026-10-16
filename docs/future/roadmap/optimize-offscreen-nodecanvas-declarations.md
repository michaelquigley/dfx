---
title: Optimize offscreen NodeCanvas declarations
state: researching
created: 2026-10-15
tags: [enhancement, spike]
source: nodegraph mini-map stage 1 baseline
milestone: v0.1.x
---

If the 5,000-node stress example remains sluggish after the direct text-color lookup fix, profile and design a spatial visibility path that avoids submitting offscreen node widgets while preserving current-detent node bounds, pin geometry, links, fit measurement, drag previews, and input routing. The application must still declare its graph every frame; do not silently make last-frame geometry authoritative. Compare both stress layouts and link densities before and after. Resolve the geometry contract before implementing this as the mini-map's performance prerequisite.
