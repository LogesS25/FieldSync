# Social Work Field Practicum Platform — Business Requirements

> **Status:** This document reflects the latest stakeholder discussion.
>
> **Important:** Items explicitly marked **TBD / To Be Provided** are not finalized and must not be hard-coded as business rules.

## 1. Purpose

The application is a digital platform connecting **universities, students, faculty supervisors, and agency supervisors** for managing the complete Social Work fieldwork practicum.

Universities are the primary institutional partners. A university has tie-ups with agencies, maintains its agency and faculty-supervisor information, defines basic practicum requirements, and manages university-specific manuals/guidelines.

The platform is **not only an activity-recording application**. It manages the complete practicum workflow:

```text
University
    ↓
Agencies + Faculty Supervisors
    ↓
Student selects fieldwork + supervisors
    ↓
Agency Supervisor + Faculty Supervisor accept
    ↓
3-person practicum team
    ↓
Daily fieldwork
    ├── Morning attendance → Agency approval → Faculty approval
    ├── Fieldwork
    ├── Handwritten daily report
    ├── Competency / fieldwork requirements
    │       → Agency approval/rejection
    │       → Faculty approval/rejection
    ├── Feedback
    └── Evening attendance → Agency approval → Faculty approval
    ↓
Consolidated Report
    → Agency review
    → Faculty review
    ↓
Evaluation
    → Agency marks + Faculty marks
    ↓
System checks university requirements
    ↓
Student progress / practicum completion
```

---

# 2. Core Business Structure

The main relationship is:

```text
UNIVERSITY
    ↓
AGENCIES + FACULTY SUPERVISORS
    ↓
STUDENT
    ↓
FIELDWORK
    ↓
AGENCY SUPERVISOR + FACULTY SUPERVISOR
```

For a student's fieldwork, the following three people form the working practicum team:

- Student
- Agency Supervisor
- Faculty Supervisor

---

# 3. Stakeholders

## 3.1 University

The university is the institutional partner.

Responsibilities:

- Register with the platform.
- Provide university-specific basic practicum requirements.
- Provide/upload the list of agencies tied to the university.
- Provide/upload the list of faculty supervisors.
- Upload/update university-specific manuals/guidelines.

## 3.2 Student

The student performs the fieldwork practicum.

Responsibilities:

- Log in.
- Select an agency.
- Select a fieldwork component from the university-provided list.
- Select the fieldwork.
- Select a faculty supervisor.
- Select an agency supervisor.
- Record morning attendance.
- Record evening attendance.
- Upload daily handwritten reports.
- Check/select applicable competency or fieldwork requirements.
- Submit consolidated report.
- View progress and requirement status.
- Receive feedback from supervisors.

## 3.3 Agency Supervisor

The agency supervisor is responsible for field-level supervision.

Responsibilities:

- Receive student's supervisor/team request.
- Accept/reject the request.
- Approve/reject morning attendance.
- Approve/reject evening attendance.
- Approve/reject competency/fieldwork requirements.
- Review consolidated report.
- Approve/reject consolidated report.
- Provide feedback.
- Provide evaluation marks.

## 3.4 Faculty Supervisor

The faculty supervisor provides academic supervision.

Responsibilities:

- Receive student's supervisor/team request.
- Accept/reject the request.
- Approve/reject morning attendance.
- Approve/reject evening attendance.
- Approve/reject competency/fieldwork requirements.
- Review consolidated report.
- Approve/reject consolidated report.
- Provide feedback.
- Provide evaluation marks.

---

# 4. University Registration and Setup

When a university registers, the application should collect the university's basic practicum requirements.

These requirements can include:

- Attendance requirements.
- Checkbox-based requirements.
- Other basic practicum conditions.
- Agency list.
- Faculty supervisor list.
- Fieldwork component list.
- University-specific manuals/guidelines.

## University Control

The university has full control over its own requirements and lists.

The university can:

- Add items.
- Remove items.
- Modify existing items.
- Make changes at any time.

This applies to all university-defined requirements and lists, including agencies, faculty supervisors, fieldwork components, attendance requirements, checkbox conditions, manuals/guidelines, and any other university-defined lists.

Students and agency supervisors must **not** be able to modify these university-defined requirements/lists.

## Important

The exact requirements have not yet been provided.

**Do not assume fixed values.**

In particular:

> **The exact minimum/base attendance percentage is TBD / To Be Provided.**

The application should therefore support university-specific/configurable requirements.

---

# 5. University Agency Management

A university will have tie-ups with agencies.

The university will upload/provide its agency list.

Business flow:

```text
University registers
        ↓
University provides agency list
        ↓
Agencies become available for that university
        ↓
Students can select from those agencies
```

A student should only see/select agencies available to their university.

---

# 6. Faculty Supervisor Management

The university will upload/provide its faculty supervisor list.

Faculty supervisors are associated with the university and should be available to students when selecting a faculty supervisor.

---

# 7. Student Onboarding and Fieldwork Selection

After student login, the student needs to:

1. Select an agency from the university's available agency list.
2. Select a fieldwork component from the university-provided list.
3. Select the relevant fieldwork.
4. Select a faculty supervisor.
5. Select an agency supervisor.

The university provides the list of fieldwork components; the student selects from it rather than providing/uploading one.

Do not invent additional business rules around fieldwork components/selection beyond what the stakeholder has confirmed here.

---

# 8. Team Formation and Acceptance

After the student selects the agency supervisor and faculty supervisor:

```text
Student selects:
    ├── Agency
    ├── Fieldwork
    ├── Faculty Supervisor
    └── Agency Supervisor

            ↓

Agency Supervisor receives notification
            ↓
        Accept / Reject

Faculty Supervisor receives notification
            ↓
        Accept / Reject

            ↓

Student + Agency Supervisor + Faculty Supervisor
            ↓
        Practicum Team
```

The supervisors must receive a notification.

The exact rule for whether **both supervisors must accept before fieldwork can start** should be confirmed.

---

# 9. Daily Fieldwork Workflow

Each fieldwork day follows a structured process.

The stakeholder specifically defined **morning and evening attendance**.

## 9.1 Morning Attendance

```text
Student records morning attendance
            ↓
Agency Supervisor approves
            ↓
Faculty Supervisor approves
```

Both supervisors participate in the approval flow.

## 9.2 Fieldwork

The student goes to the fieldwork placement and performs the day's fieldwork.

## 9.3 Evening Attendance

```text
Student records evening attendance
            ↓
Agency Supervisor approves
            ↓
Faculty Supervisor approves
```

## 9.4 End of Day

Conceptually:

```text
Morning Attendance
        ↓
Agency Approval
        ↓
Faculty Approval
        ↓
Fieldwork
        ↓
Daily Report
        ↓
Competency Requirements
        ↓
Feedback
        ↓
Evening Attendance
        ↓
Agency Approval
        ↓
Faculty Approval
        ↓
Day Complete
```

---

# 10. Daily Handwritten Report

After the day's fieldwork, the student uploads a handwritten report for that day.

Business requirements:

- Student uploads the handwritten daily fieldwork report.
- The report becomes part of the student's practicum record.
- After submission, the agency supervisor and faculty supervisor are notified.
- The agency supervisor reviews and approves/rejects the submission.
- The faculty supervisor reviews and approves/rejects the submission.

Approval flow:

```text
Student submits report
            ↓
Agency Supervisor reviews
    Approve / Reject
            ↓
Faculty Supervisor reviews
    Approve / Reject
```

If the report is rejected, the exact correction/resubmission workflow is still TBD unless already defined elsewhere in this document.

---

# 11. Competency / Fieldwork Requirements

The practicum contains competency/fieldwork requirements.

For relevant fieldwork, the student checks/selects the requirements that were addressed or completed.

Approval flow:

```text
Student checks competency / requirement
            ↓
Agency Supervisor
    Approve / Reject
            ↓
Faculty Supervisor
    Approve / Reject
```

The exact competency list has not yet been provided.

The exact correction/resubmission workflow after rejection has also not yet been provided.

---

# 12. Feedback

Both supervisors provide feedback about the student's fieldwork.

Feedback is provided by:

- Agency Supervisor
- Faculty Supervisor

Both the agency supervisor and faculty supervisor must provide feedback every weekend. Weekly feedback is mandatory for both supervisors.

The feedback should be associated with the student's fieldwork/practicum record.

The exact rule for whether feedback is attached to specific activities/reports (versus the practicum record generally) is still TBD.

---

# 13. Consolidated Report

Students must submit a consolidated report covering their fieldwork.

The university sets the submission deadline. The application does not require a predefined report format.

The agency supervisor gets the opportunity to approve or reject the report. The faculty supervisor gets the opportunity to approve or reject the report.

Approval flow:

```text
Student submits consolidated report
            ↓
Agency Supervisor reviews
    Approve / Reject
            ↓
Faculty Supervisor reviews
    Approve / Reject
```

If the report is rejected, the student must resubmit it. The resubmitted report goes through the approval process again.

---

# 14. Evaluation

The application contains an evaluation process.

Both supervisors provide marks:

```text
Agency Supervisor
        ↓
    Evaluation Mark

Faculty Supervisor
        ↓
    Evaluation Mark
```

The exact evaluation criteria, rating/marks scale, weightage, and timing are **TBD**.

## To Be Provided

- Exact evaluation form.
- Evaluation criteria.
- Maximum marks.
- Scoring method.
- Agency supervisor weightage.
- Faculty supervisor weightage.
- How the two marks are combined.
- Final evaluation calculation.

Do not invent these rules.

---

# 15. Attendance and Total Fieldwork Hours

Attendance should automatically contribute to the student's total fieldwork hours.

Conceptually:

```text
Approved Attendance Records
            ↓
Calculate Fieldwork Time
            ↓
Student Total Fieldwork Hours
```

Only attendance that satisfies the university's approval/validation process should contribute to the final total.

## TBD — Hours Calculation

The following rules have not yet been provided:

- How morning/evening attendance translates to hours.
- Break handling.
- Partial attendance.
- Late arrival.
- Early departure.
- Missing one half of a day.
- Holidays.
- Corrections.
- Whether rejected attendance contributes to hours.

Do not hard-code assumptions for these rules.

---

# 16. Basic Requirement Checking

The system should automatically check whether the student meets the basic practicum requirements defined by their university.

Possible requirements include:

- Minimum attendance percentage.
- Required fieldwork hours.
- Required fieldwork/competency completion.
- Required reports.
- Other university-defined checkbox/conditions.

The system should be able to identify whether a student:

- Meets the requirement.
- Has not yet met the requirement.
- Has missing information.
- Has pending approvals.

## Critical TBD

The exact minimum attendance percentage has **not yet been provided**.

Do not assume:

- 75%
- 80%
- 85%
- Any other fixed value

The requirement must be configurable and supplied by the university.

---

# 17. Guidelines and Manuals

The application should contain a **Guidelines** page.

Users should be able to access the relevant practicum guidance/manual.

Universities should have the ability to:

- Upload a new manual.
- Update the manual.
- Make the updated manual available in the application.

The exact rules for:

- Manual versioning
- Replacing old manuals
- Archiving old manuals
- Which users can see which manual

are still TBD.

---

# 18. Complete Business Flow

```text
UNIVERSITY REGISTRATION
        ↓
University provides:
    ├── Basic requirements
    ├── Attendance requirement
    ├── Agency list
    ├── Faculty supervisor list
    ├── Fieldwork component list
    └── Manual / Guidelines
        ↓
STUDENT LOGIN
        ↓
Student selects:
    ├── Agency
    ├── Fieldwork component
    ├── Fieldwork
    ├── Faculty Supervisor
    └── Agency Supervisor
        ↓
SUPERVISORS RECEIVE NOTIFICATION
        ↓
Agency Supervisor accepts/rejects
Faculty Supervisor accepts/rejects
        ↓
PRACTICUM TEAM
Student + Agency Supervisor + Faculty Supervisor
        ↓
DAILY FIELDWORK
        ↓
Morning Attendance
    → Agency Approval
    → Faculty Approval
        ↓
Fieldwork
        ↓
Daily Handwritten Report
        ↓
Student Checks Competencies
    → Agency Approval/Rejection
    → Faculty Approval/Rejection
        ↓
Feedback
        ↓
Evening Attendance
    → Agency Approval
    → Faculty Approval
        ↓
DAY COMPLETE
        ↓
... repeat for fieldwork period ...
        ↓
CONSOLIDATED REPORT
    → Agency Review
    → Faculty Review
        ↓
EVALUATION
    → Agency Marks
    → Faculty Marks
        ↓
SYSTEM CHECKS BASIC REQUIREMENTS
    → Attendance
    → Fieldwork Hours
    → Competencies
    → Reports
    → University-specific requirements
        ↓
STUDENT PROGRESS / PRACTICUM COMPLETION
```

---

# 19. Business States

The application should represent meaningful states for workflow records.

These are conceptual states and should be refined with stakeholders.

## Supervisor Request

```text
Pending
   ↓
Accepted / Rejected
```

## Attendance

```text
Submitted
   ↓
Agency Approved
   ↓
Faculty Approved
```

Rejection should be supported.

## Competency

```text
Selected
   ↓
Agency Approved / Rejected
   ↓
Faculty Approved / Rejected
```

## Consolidated Report

```text
Submitted
   ↓
Agency Approved / Rejected
   ↓
Faculty Approved / Rejected
```

## Evaluation

```text
Pending
   ↓
Agency Marked
+
Faculty Marked
   ↓
Complete
```

## Manual

Conceptually:

```text
Current
   ↓
Updated
   ↓
Potentially Archived
```

Exact manual lifecycle is TBD.

---

# 20. Items Still To Be Provided

The following requirements are not finalized:

- Exact minimum attendance percentage.
- Exact university basic requirements.
- Exact checkbox conditions.
- Exact fieldwork components.
- Exact fieldwork categories.
- Exact competency/fieldwork requirement list.
- Exact competency rubric.
- Exact evaluation form.
- Exact evaluation marks.
- How agency and faculty marks are combined.
- Exact attendance-to-hours calculation.
- Consolidated report deadline (university-set; exact per-university value TBD).
- Supervision frequency and exact supervision workflow, if applicable.
- Manual versioning/replacement rules.
- Correction/resubmission rules after rejection of the daily handwritten report (consolidated report resubmission is defined — see §13).

---

# 21. Important Stakeholder Decisions

Before treating the workflows as final, clarify:

1. Can university requirements differ between universities?
2. Can a student change the selected agency after the team is formed?
3. Can a student change supervisors after the team is formed?
4. Must both supervisors accept before fieldwork can start?
5. What happens if one supervisor accepts and the other rejects?
6. Does every morning attendance require both approvals?
7. Does every evening attendance require both approvals?
8. What happens if one supervisor approves attendance and the other rejects it?
9. What happens when a competency is rejected?
10. Can students edit a rejected record?
11. Can students resubmit rejected records? (Resolved for the consolidated report — see §13 — still open for other record types.)
12. What exactly counts toward total fieldwork hours?
13. What happens if only morning or only evening attendance is recorded?
14. What exactly does the basic-requirement check determine?
15. What is the final definition of practicum completion?

---

# 22. Business Definition

> **A university-connected platform where students complete and document fieldwork with a selected agency and two supervisors, while the agency and faculty supervisors jointly verify attendance and competencies, review reports, provide feedback and marks, and the system tracks practicum progress and university-defined requirements.**

---

# 23. Rules for Product Development

This document is the current **business source of truth** based on the latest stakeholder discussion.

Claude Code / developers should follow these rules:

1. **Do not invent missing business rules.**
2. **Do not hard-code TBD values.**
3. University-specific requirements must be configurable.
4. The exact attendance percentage is TBD and must not be assumed.
5. Exact competency criteria are TBD until provided.
6. Exact evaluation scoring is TBD until provided.
7. Exact report formats are TBD until provided, except where explicitly resolved in this document (the consolidated report has no predefined format — see §13).
8. Where the workflow is ambiguous, flag the ambiguity instead of silently deciding it.
9. Keep business logic separate from UI/technical implementation.
10. When implementing a requirement, refer back to this document to ensure the implementation matches the stakeholder-approved workflow.

---

# 24. Current Product Priority

The immediate product priority is the **mobile application workflow** for:

- Students
- Agency Supervisors
- Faculty Supervisors

The administrator experience can be developed later as a web dashboard.

The initial implementation should focus on the core practicum workflow described in this document, while keeping university-specific requirements and TBD business rules configurable.
