# SMO Login — UI Kit

The login page for SMO organizers. **Mobile-first**, dark by default, two fields and one CTA. No social, no "remember me", no "forgot password" at MVP.

## Files

| File | What |
|---|---|
| `index.html` | Live form on top, states gallery below (default · focus · error · loading) |
| `login.css` | Component styles — uses tokens from `../../colors_and_type.css` only |
| `Icons.jsx` | Lucide-shaped icons for the page (eye, eye-off, alert-circle, loader-2). Swap for `lucide-vue-next` in production. |
| `Wordmark.jsx` | The "SMO·" wordmark (off-white, blue dot accent). |
| `TextField.jsx` | Label + filled input. No border in default state. 2px inset focus ring. |
| `PrimaryButton.jsx` | Full-width CTA. Loading state replaces label with spinner. |
| `InlineError.jsx` | Below-form error chip. Never a toast. |
| `LoginForm.jsx` | The whole form. Holds state and submission. Accepts `forcedState` to pin a state for the gallery. |

## States

- **Default** — email autofocused, fields empty.
- **Focus** — 2px inset ring in `--action-primary` on the active field.
- **Error** — both fields get a red 2px inset ring; inline message below the form.
- **Loading** — button disabled, label replaced by spinner + "Connexion…".

## Production parity (Nuxt 3 + Vue 3 + Tailwind)

The JSX components map 1:1 to Vue SFCs. To port:

1. Lift the values out of `colors_and_type.css` into `tailwind.config.ts` (`theme.extend.colors`, `fontFamily`, `borderRadius`, `boxShadow`).
2. Convert each `.jsx` to a `.vue` SFC — same class names, same DOM. The CSS file can stay as-is or be split into Tailwind `@apply` rules.
3. Replace the local `Icons.jsx` exports with `import { Eye, EyeOff, AlertCircle, Loader2 } from 'lucide-vue-next'`.
4. Replace the fake `setTimeout` submission in `LoginForm` with a real `useFetch('/api/auth/login')` call; on error set `error.value = 'Identifiants incorrects.'`.

## Anti-patterns we deliberately avoid

- ❌ Floating card with shadow over a complex page. The form sits directly on `--bg-base`.
- ❌ Stripe-style hero illustration on desktop left.
- ❌ More than 2 type levels (display + body; caption only for the footer note).
- ❌ Toast notifications for credential errors.
- ❌ Page-level loaders.
