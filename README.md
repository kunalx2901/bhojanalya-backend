🍽️ Bhojanalya – Restaurant Onboarding Automation System

Digitizing & automating restaurant onboarding with workflow intelligence, approvals, and AI-powered validation

📌 Overview

Bhojanalya is a full-stack web application that automates the restaurant onboarding process, replacing slow and error-prone manual workflows used by operations, compliance, and business teams.

The platform enables restaurants to onboard digitally, tracks approval stages, validates data intelligently, and maintains complete audit logs — all in one centralized system.

🎯 Problem Statement

Restaurant onboarding today is:

❌ Manual & time-consuming

❌ Spread across emails, spreadsheets & calls

❌ Lacks transparency & auditability

❌ Prone to missing documents & delays

✅ Our Solution

Bhojanalya provides:

✅ Structured onboarding flow

✅ Role-based access & approvals

✅ Status tracking & audit logs

✅ AI-assisted data validation

✅ Scalable backend architecture

🧠 Key Features
🔐 Authentication & Authorization

JWT-based authentication

Role-based access (Admin, Ops, Restaurant)

🏪 Restaurant Onboarding

Create & manage restaurant profiles

Upload and validate onboarding details

Draft → Review → Approved workflow

📋 Approval Workflow

Checklist-based approvals

Status transitions with logs

Multi-team collaboration

🧾 Audit Logs

Track every status change

Complete onboarding history

🤖 AI Enhancements

Auto-check onboarding completeness

Risk scoring for restaurant data

Smart validation using AI models

🛠️ Tech Stack
Backend

Node.js

Express.js

Prisma ORM

PostgreSQL

JWT Authentication

Frontend

React.js

TypeScript

Modern component architecture

DevOps & Tools

GitHub

GitHub Actions (CI/CD)

ESLint + Prettier

AI

OpenAI / Gemini APIs (Prompt-based validation)

🏗️ System Architecture
Frontend (React)
       ↓
Backend API (Node + Express)
       ↓
Prisma ORM
       ↓
PostgreSQL Database
       ↓
AI Validation Services

📂 Project Structure
bhojanalya-restaurant-onboarding/
├── backend/
│   ├── src/
│   │   ├── controllers/
│   │   ├── routes/
│   │   ├── services/
│   │   ├── middlewares/
│   │   ├── utils/
│   │   └── app.ts
│   ├── prisma/
│   └── package.json
│
├── frontend/
│   ├── src/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── services/
│   │   └── App.tsx
│   └── package.json
│
├── .github/
│   └── workflows/
├── README.md

🚀 Getting Started
1️⃣ Clone the Repository
git clone https://github.com/your-username/bhojanalya-restaurant-onboarding.git
cd bhojanalya-restaurant-onboarding

2️⃣ Backend Setup
cd backend
npm install
npx prisma migrate dev
npm run dev

3️⃣ Frontend Setup
cd frontend
npm install
npm run dev

🔁 Workflow States
DRAFT → UNDER_REVIEW → APPROVED / REJECTED


Each transition is:

Logged

Audited

Permission-controlled

👥 Team Structure

Backend Developer 1 – Auth, onboarding APIs

Backend Developer 2 – Workflow, approvals, logs

Frontend Developer – UI, forms, dashboards

📅 Development Timeline

Days 1–2: Planning, DB schema, repo setup

Days 3–7: Backend core APIs

Days 8–9: Frontend MVP

Days 10–13: AI features & CI/CD

Days 14–15: Testing & demo prep

🌟 Future Enhancements

📊 Analytics dashboard

📄 OCR document verification

🔔 Real-time notifications

🧠 Advanced ML risk models

📱 Mobile-friendly UI

📄 License

This project is open-source and available under the MIT License.
