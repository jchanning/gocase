## Description
Add OAuth2 support for single sign-on with popular providers.

## Providers to Support
- [ ] Google OAuth
- [ ] Microsoft/Azure AD
- [ ] GitHub
- [ ] Apple Sign In (optional)

## Features
- Allow users to sign up/login with OAuth providers
- Link multiple OAuth providers to same account
- Maintain existing email/password authentication as fallback
- Auto-populate user profile from OAuth data
- Secure token storage and refresh

## Technical Requirements
- Use `golang.org/x/oauth2` package
- Store OAuth tokens securely (encrypted)
- Implement proper CSRF protection
- Handle OAuth callback routes
- Update database schema for OAuth provider associations

## Priority
Medium

## Category
Feature - Authentication

## Labels
enhancement, feature, authentication, oauth
