# Feature Specification Template

Use this template for all new feature work. This is the implementation contract, not the original prompt.

---

## 1. Feature Name

## 2. Goal
What user problem does this solve?

## 3. Scope
What is included in this feature?

## 4. Non-Scope
What is explicitly not included?

## 5. Affected Users / Roles
- student
- teacher
- admin

## 6. Constraints
- technical constraints
- security constraints
- UX constraints
- performance constraints

## 7. Data Model Impact
- tables/columns affected
- models affected
- migrations required

## 8. Interface / Route Impact
- routes added or changed
- forms/query params
- template changes

## 9. Business Rules / Invariants
List the rules that must remain true.

## 10. Edge Cases
- invalid input
- unauthorized access
- empty state
- timeout / error states

## 11. Acceptance Tests
1. Given ... when ... then ...
2. Given ... when ... then ...
3. Given ... when ... then ...

## 12. TDD Plan
- test files to create/update
- red phase first
- minimum implementation
- refactor follow-up

## 13. Documentation Updates
- `docs/IMPLEMENTATION_STATUS.md`
- `docs/API.md`
- `docs/ARCHITECTURE.md` if needed
- changelog if user-visible

## 14. Rollout Notes
- deployment considerations
- config/env impact
- backward compatibility