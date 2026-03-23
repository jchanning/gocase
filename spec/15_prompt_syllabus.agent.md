1. This specification covers the addition of a Syllabus feature and a calendar-based Revision Planner for students. These are two connected capabilities: the syllabus defines what needs to be studied for a given subject and exam standard, and the revision planner uses that syllabus to build a personalised study schedule for a student working towards a specific exam date.

2. A Syllabus is a structured, authoritative list of topics that a student is expected to cover for a given subject and exam level (e.g. GCSE Mathematics, A-Level Biology). Syllabi are created and maintained by teachers and admins. Students do not create syllabi; they select from published syllabi when building a revision plan.

3. Each Syllabus belongs to an existing Subject (using the current subjects table) and is scoped to an exam standard (e.g. GCSE, A-Level). A syllabus has a title, a description, a published flag, and a record of who created it.

4. Syllabi use a two-level structure: a Syllabus contains Sections, and each Section contains Topics. For example, a GCSE Mathematics syllabus might have sections for Algebra, Geometry, Statistics, and Number, each with their own list of topics such as "Quadratic Equations" or "Pythagoras' Theorem". Sections and their topics carry an ordering field so they can be arranged in the correct curriculum sequence.

5. Each Syllabus Topic has a title, an optional description, an estimated study hours value (defaulting to 1 hour), an ordering within its section, and an optional notes field for short written study guidance. The estimated hours value is used by the scheduling algorithm to distribute topics across available study days appropriately, giving heavier topics more time than lighter ones.

6. Syllabus Topics are linked to existing Tests via a join table (syllabus_topic_tests). A single syllabus topic may be linked to multiple tests, and a test may be linked to multiple syllabus topics. When a student is studying a topic in their revision plan, the linked tests are surfaced so the student can practice with relevant multiple-choice questions. This makes the connection between curriculum content and assessment explicit and navigable.

7. The teacher and admin syllabus management interface is accessible from the existing Create & Manage area. It allows users to: list all syllabi with filtering by subject and exam standard; create a new syllabus and assign it to a subject and exam standard; edit a syllabus to add, reorder, or remove sections and topics inline; manage topic details including estimated hours, notes, and linked tests; and publish or unpublish a syllabus. Unpublished syllabi are not visible to students.

8. A Revision Plan is a student-created study schedule. To create a plan, the student selects a published syllabus, sets their exam date, sets how many hours per day they intend to study, and selects which days of the week they will study (e.g. Monday to Friday, or all seven days). The system uses these inputs to generate a complete schedule of Revision Sessions, one per topic block per study day, distributed from today until the exam date.

9. The scheduling algorithm works as follows. First, collect all available study days between tomorrow and the exam date that fall on the student's selected weekdays. Calculate total available study hours as the number of available days multiplied by the hours-per-day value. If the total estimated hours across all syllabus topics exceeds available hours, scale each topic's allocated hours proportionally so the plan still covers everything within the time available. Assign topics to days in syllabus order (section_order then topic_order), packing topics into each day up to the hours-per-day capacity. A single topic may span multiple consecutive days if its allocated hours exceed one day's capacity. Generate a Revision Session record for each (date, topic, hours_allocated) tuple produced by this process.

10. The Revision Plan view for a student is a monthly calendar grid. Each day in the calendar that has a study session shows the topic(s) scheduled for that day. Selecting a session opens a detail panel showing the topic title, study notes, and the linked tests for that topic. Each session has a status of scheduled, completed, or skipped. Students can mark a session complete or skip it directly. Completed sessions are shown in green, skipped in grey, and scheduled in blue to make progress visible at a glance.

11. A student may have only one active revision plan per syllabus. If a student already has a plan for a given syllabus they are taken to that plan. Students may delete a plan and recreate it if they want to change their exam date or study hours. Deleting a plan removes all its sessions.

12. The Revision Planner is accessed by students from a new Revision link in the student navigation menu. The revision home page shows two sections: the student's active plans (with progress summary showing completed vs total sessions and days until exam), and a list of available published syllabi the student has not yet planned for, each showing subject, exam standard, and topic count.

13. The following new database tables are required.

    syllabi: id, subject_id (FK subjects ON DELETE SET NULL), exam_standard (same CHECK as tests), title VARCHAR(255) NOT NULL, description TEXT, is_published BOOLEAN NOT NULL DEFAULT FALSE, created_by (FK users ON DELETE SET NULL), created_at, updated_at. UNIQUE(subject_id, exam_standard, title).

    syllabus_sections: id, syllabus_id (FK syllabi ON DELETE CASCADE), title VARCHAR(255) NOT NULL, section_order INTEGER NOT NULL DEFAULT 0, created_at.

    syllabus_topics: id, syllabus_id (FK syllabi ON DELETE CASCADE), section_id (FK syllabus_sections ON DELETE SET NULL, nullable), title VARCHAR(255) NOT NULL, description TEXT, estimated_hours REAL NOT NULL DEFAULT 1.0, topic_order INTEGER NOT NULL DEFAULT 0, notes_content TEXT, created_at.

    syllabus_topic_tests: syllabus_topic_id (FK syllabus_topics ON DELETE CASCADE) + test_id (FK tests ON DELETE CASCADE), composite PRIMARY KEY, added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP.

    revision_plans: id, user_id (FK users ON DELETE CASCADE), syllabus_id (FK syllabi ON DELETE CASCADE), exam_date DATE NOT NULL, hours_per_day REAL NOT NULL DEFAULT 2.0, study_days TEXT NOT NULL DEFAULT '[1,2,3,4,5]' (JSON array of integers 0=Sun through 6=Sat), created_at, updated_at. UNIQUE(user_id, syllabus_id).

    revision_sessions: id, plan_id (FK revision_plans ON DELETE CASCADE), session_date DATE NOT NULL, syllabus_topic_id (FK syllabus_topics ON DELETE CASCADE), hours_allocated REAL NOT NULL, status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled','completed','skipped')), notes TEXT, completed_at TIMESTAMP, created_at.

14. New Go model structs required in internal/models/models.go: Syllabus (with *Subject, []SyllabusSection, TopicCount computed fields), SyllabusSection (with []SyllabusTopic), SyllabusTopic (with []Test linked tests), RevisionPlan (with *Syllabus, []RevisionSession, DaysUntilExam/TotalSessions/CompletedCount computed fields, StudyDays []int serialised as JSON TEXT in DB), RevisionSession (with *SyllabusTopic).

15. Two new repository files are required.

    internal/repository/syllabus_repository.go: SyllabusRepository with methods GetAll, GetPublished, GetByID (loads sections and topics), Create, Update, Delete, Publish, Unpublish, AddSection, UpdateSection, DeleteSection, AddTopic, UpdateTopic, DeleteTopic, LinkTest, UnlinkTest. LinkTest and UnlinkTest operate on the syllabus_topic_tests join table.

    internal/repository/revision_repository.go: RevisionRepository with methods GetPlansByUser (with syllabus join and session counts), GetPlanByUserAndSyllabus (existence check), GetPlanByID (full plan with sessions and topic details), CreatePlan (transactional: insert plan then bulk-insert sessions), DeletePlan, UpdateSessionStatus (updates status and sets completed_at when status is 'completed').

16. Two new handler files are required.

    internal/handlers/syllabus_handler.go: SyllabusHandler provides admin and teacher routes for all syllabus CRUD. Uses interface-based dependency injection matching the existing handler pattern. Redirects back to the edit page after section/topic mutations.

    internal/handlers/revision_handler.go: RevisionHandler provides student routes. ShowRevision shows active plans and available syllabi. ShowCreatePlan renders the plan form with the selected syllabus summary. CreatePlan runs the scheduling algorithm server-side, validates inputs (exam date at least 2 days in future, hours_per_day 0.5–12, at least one study day selected), then calls RevisionRepository.CreatePlan in a transaction. ShowPlan prepares calendar data organised by month and week for the template. CompleteSession and SkipSession call UpdateSessionStatus. DeletePlan calls RevisionRepository.DeletePlan.

17. New routes to add to internal/server/server.go. Admin and teacher group (RequireRole("admin","teacher")): GET and POST /admin/syllabus, GET /admin/syllabus/new, GET and POST /admin/syllabus/{id}, POST /admin/syllabus/{id}/publish, POST /admin/syllabus/{id}/unpublish, POST /admin/syllabus/{id}/section, POST /admin/syllabus/{id}/section/{sid}/update, POST /admin/syllabus/{id}/section/{sid}/delete, POST /admin/syllabus/{id}/topic, POST /admin/syllabus/{id}/topic/{tid}/update, POST /admin/syllabus/{id}/topic/{tid}/delete, POST /admin/syllabus/{id}/topic/{tid}/tests, POST /admin/syllabus/{id}/topic/{tid}/tests/{testid}/delete. Authenticated group (all roles): GET /revision, GET /revision/plan/new, POST /revision/plan, GET /revision/plan/{id}, POST /revision/session/{id}/complete, POST /revision/session/{id}/skip, POST /revision/plan/{id}/delete.

18. Five new view templates are required: views/syllabus_manage.html (admin/teacher list with subject and exam_standard filters, publish badge, topic count, edit and publish/unpublish buttons), views/syllabus_edit.html (edit metadata form plus inline section list; each section expands to show its topics with estimated hours, notes, and linked-tests panel; inline forms for adding sections and topics; test search/link panel per topic), views/revision_planner.html (student home: active plans as cards with progress bar and days-until-exam countdown; published unplanned syllabi as cards with Start Planning button), views/revision_create.html (plan creation form: shows selected syllabus summary with total estimated hours; inputs for exam_date, hours_per_day, and study_days checkboxes; preview of how many available days will be generated), views/revision_plan.html (monthly calendar grid; nav arrows for previous/next month; each day cell shows session topic chips colour-coded by status; clicking a chip shows an inline detail card with topic notes and linked test links; Complete and Skip buttons per session).

19. Navigation changes: add Revision link to the student nav (both desktop and mobile) in views/layout.html, pointing to /revision. Add Syllabi link to the teacher nav pointing to /admin/syllabus. Add Syllabi link to the admin nav pointing to /admin/syllabus.

20. The existing Topic entity (tests.topic_id) remains separate from SyllabusTopic. When linking tests to a SyllabusTopic, the test search in the edit interface should default to filtering by the syllabus's subject to surface relevant tests quickly, but any published test may be linked.
