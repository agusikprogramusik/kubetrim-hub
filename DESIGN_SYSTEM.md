# KubeTrim Hub — UI/UX Design Philosophy & Prompt Specification (October 2026)

This document preserves the canonical design philosophy, negative constraints, visual direction, and two-step design-to-code workflow for **KubeTrim Hub**.

---

## Original Prompt & Specification (Verbatim)

### ROLE & DESIGN PHILOSOPHY (OCTOBER 2026)
Act as an Avant-Garde Staff Product Designer and Creative Technologist (Awwwards / Linear / Teenage Engineering / Stripe Press caliber). Your task is to design and code a high-craft UI that feels unmistakably human-designed in late 2026, completely devoid of generic "AI-generated template" aesthetics.

### STRICT NEGATIVE CONSTRAINTS (THE "ANTI-AI SLOP" RULES)
1. NO generic SaaS gradients (strictly ban indigo-to-purple, blue-to-cyan, or neon glowing blobs).
2. NO symmetrical 3-column feature cards with a centered icon in a colored circle.
3. NO centered Hero section with a generic pill-shaped badge above an H1 and two centered rounded-full buttons.
4. NO default Tailwind drop shadows (`shadow-md`, `shadow-lg`, `shadow-xl`) on cards.
5. NO pure `#FFFFFF` or `#000000` backgrounds, and NO standard Tailwind `blue-500` or `indigo-600` primary accents.
6. NO generic placeholder text like "Supercharge your workflow" or "Unlock the power of...". Write realistic, sharp, domain-specific microcopy.

### 2026 VISUAL & TECHNICAL DIRECTION (CHOOSE ONE OR BLEND)
Apply a "Tactile Editorial & Precision UI" aesthetic:
- **Typography-First Hierarchy:** Import and pair 3 distinct Google Fonts: a high-contrast Editorial Serif (e.g., 'Instrument Serif' or 'Newsreader') for dramatic headlines, a crisp geometric Sans (e.g., 'Plus Jakarta Sans' or 'Inter') for UI, and a Monospace font ('JetBrains Mono') for labels, indices (`01 //`), timestamps, and data metrics. Use extreme scale contrast (`clamp()` for fluid display type, tight tracking `-0.03em` on headings).
- **Layout & Spatial Architecture:** Use an asymmetric Swiss-style grid or asymmetric Bento architecture. Expose structural hairlines (`1px solid` borders with subtle opacity) so the layout feels engineered like a physical blueprint or a high-end print magazine. Use generous, intentional negative space.
- **Tactile Materiality:** Add a subtle fixed SVG fractalNoise overlay (`opacity: 0.035`, `pointer-events: none`) across the viewport for analog warmth. On interactive cards and surfaces, use ultra-thin `1px` borders with subtle top-edge inner highlights (`box-shadow: inset 0 1px 0 0 rgba(255,255,255,0.1)`) instead of blurry outer shadows.
- **Color Palette:** 
  - [Wybierz jedno: WARM ANALOG LIGHT: Warm alabaster/paper `#F6F5F0` background, deep charcoal `#141413` ink, muted stone `#E6E4DD` borders, and international orange `#FF4F00` or forest green `#1B4332` accent] 
  - [LUB: OBSIDIAN INDUSTRIAL DARK: Deep carbon `#0B0C0E` background, elevated surface `#121316`, hairline border `#22242A`, crisp off-white `#ECECEE` text, and raw amber `#F59E0B` or acid lime `#D4FF00` accent].
- **Micro-interactions:** Use custom cubic-bezier transitions (`cubic-bezier(0.22, 1, 0.36, 1)`). Interactive elements should react with crisp mechanical precision (e.g., subtle border brightening, monospace arrow shift `[ -> ]`, or background fill wipe).

### PROJECT BRIEF
[Tutaj wpisz po polsku lub angielsku, co dokładnie chcesz zbudować, np.: Zbuduj dashboard dla narzędzia CLI audytującego klastry Kubernetes / stronę główną dla studia fotograficznego / aplikację smart home...]

Krok 1 (Architektura wizualna): „Nie pisz jeszcze kodu. Zaproponuj dla tego projektu unikalny system designu na październik 2026 (paleta HEX, para fontów, siatka układu i 3 nietypowe detale interakcji), unikając całkowicie klisz AI/SaaS.”

Krok 2 (Implementacja): „Świetnie. Teraz zaimplementuj ten dokładnie opisany design w jednym kompletnym pliku [HTML+Tailwind / React], dbając o pełne odwzorowanie detali CSS, mikrointerakcji i typografii.”
