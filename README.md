# 🍽️ Bhojanalya Partner Portal

A comprehensive B2B web application designed to seamlessly onboard restaurant partners, process their menus via AI-driven OCR, and empower them to create data-driven promotional deals.

## ✨ Key Features

### For Restaurant Owners 🧑‍🍳

* **Multi-Step Onboarding:** A smooth, 6-step registration process capturing restaurant details, legal compliance (FSSAI/GSTIN), and operational hours.
* **Smart Menu Uploads:** Direct upload of Menu PDFs/Images, seamlessly tied to the restaurant's profile.
* **AI-Powered Deals Dashboard:**
* View market insights (Average vs. Median Cost for Two).
* Receive AI-generated deal suggestions tailored to their pricing tier (Premium vs. Mainstream).
* Create custom percentage-based or flat-rate discounts.


* **Live Storefront Preview:** Instantly preview how the restaurant and active deals will appear to end customers.

### For Platform Admins 🛡️

* **Role-Based Access Control:** Secure JWT-based admin routing.
* **Approval Workflow:** A dedicated dashboard to review incoming restaurant registrations and parsed OCR menu data.
* **Data Verification:** Admins can view uploaded menu files side-by-side with the AI-parsed data (pricing, categories, operational hours) before approving them to go live.

---

## 🛠️ Tech Stack

**Frontend:**

* **Framework:** [Next.js](https://nextjs.org/) (App Router)
* **Styling:** [Tailwind CSS](https://tailwindcss.com/)
* **Animations:** [Framer Motion](https://www.framer.com/motion/)
* **Icons:** [Lucide React](https://lucide.dev/)
* **Language:** TypeScript

**Backend (Integration):**

* Designed to consume a RESTful API running on port `8000`.
* **Authentication:** JWT (JSON Web Tokens) with `Bearer` strategy.

---

## 🔐 Authentication & Routing Flow

1. **Token Storage:** JWTs are stored in `localStorage` upon successful login.
2. **Interceptors/Helpers:** The `apiRequest` utility automatically appends the `Authorization: Bearer <token>` header to all requests.
3. **Role Checks:** * The `/admin` route decodes the JWT payload to ensure the `role` is `ADMIN`. If a standard user attempts access, they are redirected to `/deals`.
* Standard protected routes ping `/protected/ping` to verify token validity before rendering.



---

## 🌐 Key API Integrations

* `POST /api/restaurants` - Creates a new restaurant profile.
* `POST /api/menus/upload` - Handles `multipart/form-data` for menu uploads.
* `GET /api/restaurants/:id/deals/suggestion` - Fetches AI deal suggestions and market pricing stats.
* `POST /api/restaurants/:id/deals` - Publishes a batch of new deals.
* `GET /api/admin/menus/pending` - Fetches the queue of parsed menus awaiting admin verification.
* `POST /api/admin/restaurants/:id/approve` - Admin endpoint to push a restaurant live.

---

## 🎨 UI/UX Highlights

* **Optimistic UI:** Actions like removing a draft deal or approving a restaurant update the UI instantly before the server responds, ensuring a snappy experience.
* **Fluid Layouts:** Utilizing `AnimatePresence` and `layout` props from Framer Motion for buttery-smooth accordion expansions and list re-ordering.
* **Data Visualization:** Custom built, relative-scaled price lines to visualize a restaurant's standing against market medians and averages.
