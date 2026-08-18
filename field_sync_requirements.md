# Social Work Field Practicum Platform
## Requirements & Stakeholder Specification

> **Working title:** A Role-Based Digital Platform for Standardization and Competency-Based Management of Social Work Field Work Practicum  
> **App name:** TBD

---

## 1. Purpose

The platform is intended to provide a centralized, role-based digital system for managing social work field practicum.

The primary purpose is to standardize field practicum processes across institutions while providing structured monitoring, documentation, supervision, verification, and competency-based evaluation.

The platform should connect:

- Students
- Faculty Supervisors
- Agency Supervisors
- Administrators / Institution Staff

The system should reduce dependence on disconnected spreadsheets, emails, paper forms, and manually maintained records.

---

## 2. Problem Statement

Social work field practicum processes can vary between institutions, resulting in inconsistencies in:

- Field work manuals and guidelines
- Supervision practices
- Student monitoring
- Attendance and field-hour tracking
- Field work documentation
- Evaluation methods
- Competency assessment
- Verification of practicum activities

The proposed platform addresses these problems through a centralized system where field practicum activities, supervision, documentation, evaluation, and competency progress can be managed in a structured manner.

---

## 3. Goals

The platform should aim to:

1. Standardize field practicum processes.
2. Provide role-based access to different stakeholders.
3. Allow students to document field work activities in real time.
4. Track attendance and validated field work hours.
5. Allow faculty and agency supervisors to review and verify student activities.
6. Provide structured supervision and evaluation.
7. Track student competency development.
8. Identify missing, incomplete, or unverified practicum records.
9. Provide standardized manuals, templates, forms, and guidelines.
10. Generate verified practicum and competency reports.
11. Improve transparency and accountability between students, supervisors, and institutions.
12. Reduce administrative work associated with manual documentation.

---

# 4. Stakeholders

## 4.1 Student

The student is the primary participant in the field practicum.

### Responsibilities

- Maintain daily field work records.
- Record attendance.
- Document field activities.
- Submit required weekly reports.
- Follow standardized practicum guidelines.
- Review supervisor feedback.
- Monitor personal competency progress.
- Access required practicum resources.

### Student capabilities

- Login securely.
- View practicum details.
- Record daily field work.
- Record/view attendance.
- Submit reports.
- View report status.
- View supervisor feedback.
- View competency progress.
- Access manuals, guidelines, templates, and other resources.
- Download relevant documents.

---

## 4.2 Faculty Supervisor

The faculty supervisor represents the student's educational institution and is responsible for academic supervision and evaluation.

### Responsibilities

- Monitor assigned students.
- Review student field work logs.
- Verify attendance records.
- Review submitted reports.
- Schedule or record supervision sessions.
- Evaluate student performance.
- Provide feedback.
- Monitor competency development.
- Identify gaps or inconsistencies.
- Receive system notifications and alerts.

### Faculty capabilities

- View assigned students.
- View student practicum progress.
- Review field activity logs.
- Verify attendance.
- Review reports.
- Conduct/record supervision sessions.
- Submit evaluation scores.
- Provide feedback.
- View competency reports.
- Receive alerts for missing or inconsistent records.

---

## 4.3 Agency Supervisor

The agency supervisor supervises the student's field work at the placement organization.

### Responsibilities

- Confirm student attendance.
- Verify field work activities.
- Record supervision details.
- Evaluate student performance.
- Provide feedback.
- Follow standardized supervisory guidelines.
- Raise complaints or grievances when necessary.

### Agency Supervisor capabilities

- View assigned students.
- Record attendance.
- Submit supervision reports.
- Verify field work activities.
- Submit assessment scores.
- Provide feedback.
- Access supervision guidelines.
- Flag complaints or grievances.

### Supervisor Verification

The platform should support verification of agency supervisor credentials/qualification so that students are not supervised by unauthorized or unqualified personnel.

---

## 4.4 Administrator / Institution Staff

The administrator is responsible for managing the overall practicum process and maintaining standardization.

### Responsibilities

- Manage institutions.
- Manage users and roles.
- Assign supervisors to students.
- Configure competency rules/criteria.
- Manage standardized practicum resources.
- Monitor practicum activity.
- Monitor missing or incomplete records.
- Generate reports.
- Maintain standardized processes.

### Administrator capabilities

- Manage institutions.
- Manage students.
- Manage faculty supervisors.
- Manage agency supervisors.
- Assign supervisors.
- Configure competency criteria.
- Manage manuals and templates.
- Manage supervision guidelines.
- Monitor practicum progress.
- View alerts and verification issues.
- Generate reports.
- Perform authorized manual updates.

---

# 5. Core Platform Modules

The platform should contain the following major modules.

## 5.1 Authentication & Role-Based Access

Users should access the platform according to their assigned role.

Supported roles:

- Student
- Faculty Supervisor
- Agency Supervisor
- Administrator

Each role should have its own dashboard and permissions.

Users should only be able to access information and actions appropriate to their role.

---

## 5.2 Student Practicum Module

The student module should allow students to:

- View their practicum placement.
- Record daily field work.
- Record attendance.
- Submit weekly reports.
- View submission status.
- View supervisor feedback.
- View competency progress.
- Access standardized resources.

Student activity records should be time-stamped for verification.

---

## 5.3 Faculty Supervision Module

The faculty supervision module should allow faculty supervisors to:

- View assigned students.
- Monitor student activity.
- Review field work logs.
- Verify attendance.
- Review submitted reports.
- Schedule/record supervision sessions.
- Evaluate students.
- Provide feedback.
- View competency progress.
- Receive alerts.

---

## 5.4 Agency Supervision Module

The agency supervision module should allow agency supervisors to:

- View assigned students.
- Record attendance.
- Record supervision sessions.
- Verify field work.
- Submit assessments.
- Provide feedback.
- Access supervisory guidelines.
- Raise grievances or complaints.

---

## 5.5 Competency-Based Evaluation Module

The platform should evaluate student competency using predefined criteria/rubrics.

The evaluation should consider information such as:

- Attendance
- Completion of field activities
- Validated field work hours
- Activity verification
- Agency supervisor feedback
- Faculty supervisor feedback
- Competency criteria/rubric scores

The platform should maintain competency scores and provide a view of the student's competency development over time.

### Important principle

Competency evaluation should be based on defined professional criteria rather than only subjective or informal assessment.

The exact competency criteria and scoring methodology should be configurable by the authorized administrator/institution.

---

## 5.6 Standardization Module

The platform should provide a common set of practicum materials and guidelines.

The standard resource set may include:

- Field work manual
- Evaluation forms
- Supervision guidelines
- Social work methodology formats/templates
- Agency standards
- Feedback forms
- Competency templates

The objective is to ensure that participating institutions and stakeholders use consistent practicum standards.

---

## 5.7 Resource Library

The platform should provide a centralized resource library containing materials required by students and supervisors.

Possible resources include:

- Downloadable manuals
- Training materials
- Standardized templates
- Guidelines
- Evaluation forms
- Feedback forms
- Other approved practicum resources

Resources should be accessible according to the user's role and permissions where required.

---

## 5.8 Monitoring & Verification Module

The platform should continuously check practicum records against defined requirements.

The system should be capable of identifying:

- Missing reports
- Missing supervision
- Unverified attendance
- Incomplete field work records
- Missing field hours
- Unauthorized/non-qualified supervisors
- Incomplete evaluations
- Competency progress gaps

When a mismatch or missing requirement is identified, the appropriate stakeholder should receive an alert or notification.

---

## 5.9 Reporting Module

The platform should generate structured practicum reports.

Reports may include:

- Attendance reports
- Field work/activity reports
- Supervision reports
- Student progress reports
- Competency reports
- Evaluation reports
- Verification reports
- Institution-level practicum reports

Reports should use verified data wherever applicable.

---

# 6. Core Workflow

The overall practicum workflow should follow this general process:

### Step 1 — Student Records Field Work

The student records daily practicum activities and attendance.

### Step 2 — Data Submission

The student submits the required field work records/reports.

### Step 3 — Agency Verification

The agency supervisor reviews or verifies relevant attendance and field work activities and records supervision information.

### Step 4 — Faculty Review

The faculty supervisor reviews student records, reports, attendance, and other relevant information.

### Step 5 — Evaluation

Faculty and agency supervisors provide evaluations and feedback according to the defined criteria.

### Step 6 — Competency Tracking

The system uses the available verified information to calculate/maintain competency progress according to the configured competency framework.

### Step 7 — Monitoring & Alerts

The system identifies missing, incomplete, inconsistent, or unverified information and alerts the appropriate stakeholders.

### Step 8 — Reporting

Verified practicum data is used to generate student, supervisor, and institution-level reports.

### Step 9 — Administrative Oversight

Administrators monitor the overall practicum process and ensure that standardized procedures are being followed.

---

# 7. Functional Requirements

## Authentication & Access

- **FR-01:** Users must be able to authenticate into the platform.
- **FR-02:** Users must be assigned a role.
- **FR-03:** The platform must provide role-specific dashboards.
- **FR-04:** Users must only access functionality permitted for their role.

## Student Requirements

- **FR-05:** Students must be able to record daily field work.
- **FR-06:** Students must be able to record attendance.
- **FR-07:** Students must be able to submit weekly reports.
- **FR-08:** Students must be able to view supervisor feedback.
- **FR-09:** Students must be able to monitor competency progress.
- **FR-10:** Students must be able to access standardized practicum resources.

## Faculty Supervisor Requirements

- **FR-11:** Faculty supervisors must be able to manage/view assigned students.
- **FR-12:** Faculty supervisors must be able to review field work logs.
- **FR-13:** Faculty supervisors must be able to verify attendance.
- **FR-14:** Faculty supervisors must be able to review reports.
- **FR-15:** Faculty supervisors must be able to record supervision sessions.
- **FR-16:** Faculty supervisors must be able to evaluate students.
- **FR-17:** Faculty supervisors must be able to provide feedback.
- **FR-18:** Faculty supervisors must be able to view competency progress.

## Agency Supervisor Requirements

- **FR-19:** Agency supervisors must be able to view assigned students.
- **FR-20:** Agency supervisors must be able to record attendance.
- **FR-21:** Agency supervisors must be able to record supervision details.
- **FR-22:** Agency supervisors must be able to verify field work activities.
- **FR-23:** Agency supervisors must be able to evaluate students.
- **FR-24:** Agency supervisors must be able to provide feedback.
- **FR-25:** Agency supervisors must be able to raise grievances/complaints.

## Competency Requirements

- **FR-26:** The platform must support predefined competency criteria/rubrics.
- **FR-27:** The platform must record competency-related evaluation data.
- **FR-28:** The platform must track competency progress.
- **FR-29:** The platform must generate competency scores/reports based on configured criteria.

## Standardization Requirements

- **FR-30:** The platform must provide standardized practicum guidelines.
- **FR-31:** The platform must provide standardized manuals and templates.
- **FR-32:** Authorized administrators must be able to manage standardized resources.
- **FR-33:** Users must be able to access the latest approved standardized resources.

## Monitoring & Verification Requirements

- **FR-34:** The platform must identify missing required records.
- **FR-35:** The platform must identify unverified attendance.
- **FR-36:** The platform must identify missing supervision records.
- **FR-37:** The platform must identify incomplete practicum submissions.
- **FR-38:** The platform must identify supervisor qualification/verification issues.
- **FR-39:** The platform must notify relevant users about identified issues.

## Reporting Requirements

- **FR-40:** The platform must generate practicum reports.
- **FR-41:** The platform must generate competency reports.
- **FR-42:** The platform must provide verified data for reporting.
- **FR-43:** Authorized users must be able to view/download relevant reports.

---

# 8. Data That the Platform Needs to Manage

At a high level, the platform will need to maintain information relating to:

### Users

- Student information
- Faculty supervisor information
- Agency supervisor information
- Administrator information
- User roles

### Institutions & Agencies

- Educational institutions
- Fieldwork agencies/organizations
- Student placements
- Supervisor assignments

### Practicum

- Practicum details
- Field work activities
- Attendance
- Field work hours
- Weekly reports
- Supervision sessions
- Feedback
- Evaluations

### Competencies

- Competency framework
- Competency criteria
- Rubrics
- Evaluation scores
- Competency progress

### Resources

- Manuals
- Guidelines
- Templates
- Forms
- Training resources
- Agency standards

### Verification & Monitoring

- Verification status
- Missing records
- Alerts
- Supervisor verification
- Activity confirmation

---

# 9. Standardization Principles

The platform should maintain a consistent practicum process while allowing authorized institutions to configure appropriate institution-specific information.

The core standardized elements should include:

1. Practicum guidelines
2. Field work documentation
3. Attendance tracking
4. Supervision process
5. Evaluation criteria
6. Competency assessment
7. Reporting
8. Verification

The goal is that students across different institutions are evaluated using consistent professional standards rather than completely different institutional processes.

---

# 10. Accountability & Verification

A major principle of the platform is that practicum information should be traceable and verifiable.

Examples:

- Student submits a field activity.
- Agency supervisor verifies the relevant activity.
- Faculty supervisor reviews the activity.
- Attendance is recorded and verified.
- Supervision sessions are recorded.
- Evaluations are linked to the relevant student/practicum.
- System identifies missing or inconsistent information.

Activity records should maintain timestamps and appropriate verification status.

---

# 11. Alerts & Notifications

The platform should provide alerts when important practicum requirements are not satisfied.

Examples include:

- Missing daily/weekly records
- Unverified attendance
- Missing supervision
- Missing evaluation
- Insufficient field hours
- Incomplete reports
- Supervisor verification problems
- Competency progress gaps

Notifications should be directed to the stakeholder responsible for resolving the issue.

---

# 12. Dashboard Expectations

## Student Dashboard

Should provide a clear overview of:

- Practicum status
- Attendance
- Completed field hours
- Recent field activities
- Pending submissions
- Reports
- Supervisor feedback
- Competency progress
- Alerts
- Resources

## Faculty Supervisor Dashboard

Should provide:

- Assigned students
- Student practicum status
- Pending reviews
- Attendance verification
- Reports requiring review
- Supervision schedule/status
- Evaluation status
- Competency progress
- Alerts

## Agency Supervisor Dashboard

Should provide:

- Assigned students
- Attendance
- Field activities requiring verification
- Supervision records
- Assessments
- Feedback
- Guidelines
- Alerts

## Administrator Dashboard

Should provide an overall view of:

- Institutions
- Students
- Faculty supervisors
- Agency supervisors
- Placements
- Practicum progress
- Verification issues
- Missing records
- Competency progress
- Reports
- Standardized resources

---

# 13. Out of Scope / Not Yet Defined

The source document does not define the following in enough detail and they should not be assumed during initial implementation:

- Specific technology stack
- Database technology
- Hosting/cloud provider
- Detailed API design
- Exact UI design
- Exact competency scoring formula
- Exact competency framework to be used
- Detailed institution onboarding process
- Detailed notification channels
- Payment/subscription functionality
- Detailed mobile application requirements
- Detailed privacy/data-retention policy
- Integration with external university systems

These should be decided separately during product and technical design.

---

# 14. Important Product Decisions Still Required

Before implementing the complete system, the stakeholders should clarify:

1. What exact competency framework will be used?
2. What are the exact competency categories and criteria?
3. How are competency scores calculated?
4. What constitutes a valid field work activity?
5. Who can verify each type of record?
6. What are the minimum required field hours?
7. How frequently must students submit reports?
8. How frequently must supervision occur?
9. What qualifications are required for an agency supervisor?
10. Can institutions customize practicum requirements?
11. Which elements must remain globally standardized?
12. What happens when a supervisor rejects a submission?
13. Can students edit records after submission?
14. What is the approval/rejection workflow?
15. What reports are mandatory?
16. Who can view student competency records?
17. How are grievances/complaints handled?
18. What notifications are required and through which channels?

---

# 15. MVP Scope

The first version should focus on the core practicum workflow rather than attempting to implement every possible feature.

### MVP Users

- Student
- Faculty Supervisor
- Agency Supervisor
- Administrator

### MVP Features

- Authentication and role-based access
- Student management
- Institution/agency management
- Supervisor assignment
- Practicum placement management
- Daily field work logging
- Attendance tracking
- Weekly report submission
- Agency supervisor verification
- Faculty supervisor review
- Supervision records
- Basic evaluations
- Competency tracking
- Standardized resource library
- Alerts for missing/incomplete records
- Basic practicum reports
- Administrator monitoring

Advanced features can be added after the core workflow has been validated with actual stakeholders.

---

# 16. Guiding Product Principle

The platform should function as a **single source of truth for social work field practicum**.

The system should bring together:

**Documentation + Attendance + Supervision + Verification + Evaluation + Competency Tracking + Standardized Resources + Reporting**

into one role-based platform.

The ultimate objective is to make social work field practicum more **standardized, transparent, verifiable, competency-focused, and accountable** across institutions.
