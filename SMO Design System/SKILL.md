---
name: smo-design
description: Use this skill to generate well-branded interfaces and assets for SMO (sportpotes.fr — amateur 5v5 football match organizer), either for production or throwaway prototypes/mocks/etc. Contains essential design guidelines, colors, type, fonts, assets, and UI kit components for prototyping.
user-invocable: true
---

Read the README.md file within this skill, and explore the other available files.

If creating visual artifacts (slides, mocks, throwaway prototypes, etc), copy assets out and create static HTML files for the user to view. If working on production code, you can copy assets and read the rules here to become an expert in designing with this brand.

If the user invokes this skill without any other guidance, ask them what they want to build or design, ask some questions, and act as an expert designer who outputs HTML artifacts _or_ production code, depending on the need.

## Hard constraints (do not violate)

- **Dark by default, no light mode.** Body is always `#0E1014`.
- **Strict palette** — only the colors in `colors_and_type.css`. No purple, orange, pink, warm gradients, or pure white backgrounds.
- **Two type levels max** per screen: display + body. Caption is a third only for genuine micro-copy.
- **No floating cards with shadows over complex backgrounds.** The "AI login page" cliché. Forms sit directly on the page.
- **No emoji in product UI.** Icons come from Lucide only.
- **French copy, tutoiement** (informal `tu`).
- Production stack target: Nuxt 3 + Vue 3 + Tailwind + lucide-vue-next.
