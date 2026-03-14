```chatagent
Create and implement a UI/UX modernization plan for GoCaSE so the application looks modern, professional, and suitable for education.

Phase 1: Design Foundation
1. Define a calm, education-focused visual style with strong readability and clear hierarchy.
2. Introduce a semantic color system used across all pages: Primary, Secondary, Accent, Background, Surface, Border, Text Primary, Text Secondary, Success, Warning, Error.
3. Ensure color choices meet accessibility contrast requirements (WCAG AA minimum) for text, buttons, and form controls.
4. Standardize typography scale for page titles, section headings, body text, labels, and helper text.
5. Standardize spacing and layout rhythm (consistent margins, paddings, card spacing, and section gaps).

Phase 2: Global Layout and Navigation
6. Improve top-level navigation so Student, Teacher, and Admin users can quickly find relevant pages.
7. Add consistent page headers with title, context, and primary action area.
8. Make all key pages responsive for desktop, tablet, and mobile widths.
9. Ensure consistent empty states, loading states, and error states.

Phase 3: Component Standardization
10. Standardize button variants (primary, secondary, danger, disabled, loading).
11. Standardize form fields (input, select, textarea, file upload) including labels, hints, validation, and error messages.
12. Standardize cards, tables, badges, and alert messages with consistent visual treatment.
13. Ensure all interactive elements have clear hover, focus, active, and disabled states.

Phase 4: Core Journey UX Improvements
14. Improve authentication flow (login/register) with clearer validation and user guidance.
15. Improve Student dashboard readability: progress, recent tests, scores, and study streak visibility.
16. Improve test discovery flow with clear filter controls and result count feedback.
17. Improve test-taking experience with better question readability, answer spacing, timer clarity, and confidence indicators.
18. Improve results and review pages with clearer score summaries and actionable feedback.
19. Improve Teacher workflows for create/edit/upload test tasks, including notes upload and preview clarity.
20. Improve Admin workflows for user management and test management with clearer table actions and confirmation patterns.

Phase 5: Accessibility and Quality
21. Ensure full keyboard navigation across forms, tables, dialogs, and test-taking controls.
22. Ensure visible focus outlines on all interactive elements.
23. Ensure meaningful ARIA labels and landmarks where needed.
24. Run visual QA across all main pages for consistency and regressions.

Phase 6: Measurement and Rollout
25. Define UX success criteria: reduced user errors, faster task completion, improved readability, and improved perceived trust.
26. Perform quick usability testing with representative Student, Teacher, and Admin scenarios.
27. Implement improvements in small increments and update changelog after each milestone.

Acceptance Criteria
- The UI uses a consistent semantic color and typography system across all pages.
- The application appears visually consistent, modern, and appropriate for an educational setting.
- Core Student, Teacher, and Admin workflows are easier to complete with fewer steps and clearer feedback.
- All primary pages are responsive and accessible (WCAG AA contrast and keyboard support).
```