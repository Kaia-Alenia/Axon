## 2026-06-21 - Added semantics to Compose UI
**Learning:** In Jetpack Compose, icon-only buttons using just `Box` and `.clickable` lack accessibility context for screen readers like TalkBack, announcing only 'Double tap to activate'. Adding `onClickLabel` and `role = Role.Button` to the `.clickable` modifier fixes this.
**Action:** Always verify if custom Compose buttons have proper semantics for screen readers.
