## Description
Implement an email notification system to keep users informed about important events.

## Features
- Email notification on test completion
- Grade availability notifications for students
- New test assignment notifications
- Teacher feedback notifications
- Configurable email preferences per user
- Email templates using Go html/template
- Support for SMTP configuration

## Technical Requirements
- Use Go standard library `net/smtp` or third-party library like gomail
- Store email preferences in user settings
- Queue system for batch email sending
- Email template system
- Configuration via environment variables

## Priority
High

## Category
Feature - User Experience

## Labels
enhancement, feature, email, notifications
