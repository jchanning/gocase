# Branch Protection Setup Guide

This document provides instructions for setting up branch protection rules for the GoCaSE repository.

## Recommended Branch Protection Rules

### Main Branch Protection

Navigate to: **Settings** → **Branches** → **Add branch protection rule**

#### Rule Name
`main`

#### Recommended Settings

**Protect matching branches:**
- ✅ **Require a pull request before merging**
  - ✅ Require approvals: **1**
  - ✅ Dismiss stale pull request approvals when new commits are pushed
  - ✅ Require review from Code Owners (if CODEOWNERS file exists)
  
- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - Required status checks:
    - `Test`
    - `Build`
    - `Lint`
    - `Security Scanning`

- ✅ **Require conversation resolution before merging**

- ✅ **Require signed commits** (optional but recommended)

- ✅ **Require linear history** (prevents merge commits, enforces rebase/squash)

- ✅ **Include administrators** (applies rules to admins too)

- ❌ **Allow force pushes** (keep disabled for main branch)

- ❌ **Allow deletions** (keep disabled for main branch)

### Develop Branch Protection (Optional)

If using a develop branch for ongoing development:

#### Rule Name
`develop`

#### Settings
- ✅ Require a pull request before merging
  - Require approvals: **1**
- ✅ Require status checks to pass
- ✅ Require conversation resolution
- ✅ Allow force pushes (for rebasing feature branches)
- ❌ Allow deletions

## Setting Up via GitHub UI

1. Go to your repository on GitHub
2. Click **Settings** (repository settings, not account settings)
3. Click **Branches** in the left sidebar
4. Under "Branch protection rules", click **Add rule**
5. Enter `main` as the branch name pattern
6. Configure the settings as listed above
7. Click **Create** or **Save changes**

## Setting Up via GitHub CLI

```bash
# Install GitHub CLI if not already installed
# https://cli.github.com/

# Enable branch protection for main
gh api repos/jchanning/gocase/branches/main/protection \
  --method PUT \
  --field required_status_checks[strict]=true \
  --field required_status_checks[contexts][]=Test \
  --field required_status_checks[contexts][]=Build \
  --field required_status_checks[contexts][]=Lint \
  --field required_pull_request_reviews[required_approving_review_count]=1 \
  --field required_pull_request_reviews[dismiss_stale_reviews]=true \
  --field enforce_admins=true \
  --field required_linear_history=true \
  --field allow_force_pushes=false \
  --field allow_deletions=false
```

## CODEOWNERS File (Optional)

Create a `.github/CODEOWNERS` file to automatically request reviews from specific people:

```
# Default owners for everything in the repo
* @jchanning

# Specific owners for different parts
/internal/database/ @jchanning
/views/ @jchanning
/.github/workflows/ @jchanning
```

## Rulesets (New GitHub Feature)

GitHub now offers "Rulesets" as a more flexible alternative to branch protection rules:

1. Go to **Settings** → **Rules** → **Rulesets**
2. Click **New ruleset** → **New branch ruleset**
3. Configure similar rules with more granular control
4. Rulesets can target multiple branches with patterns

## Benefits of Branch Protection

1. **Quality Assurance**: All code goes through CI/CD pipeline
2. **Code Review**: Ensures peer review before merging
3. **Prevent Accidents**: Protects against accidental deletions or force pushes
4. **Audit Trail**: Clear history of who approved what changes
5. **Consistency**: Enforces team standards and best practices

## Workflow After Protection

Once protected, contributors should:

1. Create feature branches from `main`
2. Make changes and commit
3. Push to their fork or feature branch
4. Open a pull request
5. Wait for CI checks to pass
6. Request reviews
7. Address feedback
8. Merge once approved and checks pass

## Troubleshooting

**Can't push to main:**
- This is expected! Create a feature branch instead.
- Use: `git checkout -b feature/my-feature`

**Status checks failing:**
- Check the Actions tab for details
- Fix issues and push new commits
- Checks will re-run automatically

**Need to bypass protection:**
- Only admins can bypass (if "Include administrators" is unchecked)
- Strongly discouraged except for emergencies
- Use Settings → Branches → temporarily disable protection if absolutely necessary

## References

- [GitHub Docs: Branch Protection](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches)
- [GitHub Docs: Rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets)
