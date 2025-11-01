# PLTMH Payment Module Documentation

## Overview
JavaScript module for handling electricity payment processing for both prepaid (token) and postpaid (bill) payment systems for PLTMH Lembang Palesan.

**Version:** 1.0.0  
**Author:** PLTMH Lembang Palesan Development Team

---

## Module Structure

### Global Variables
- `selectedPaymentType` (string): Currently selected payment type ('prepaid' or 'postpaid')
- `selectedAmount` (number): Amount to be paid
- `selectedPaymentMethod` (string): Selected payment method ('qr', 'bank', or 'ewallet')
- `customerData` (object|null): Customer information from backend
- `prepaidBlocked` (boolean): Flag to prevent prepaid payment processing
- `tokenOptions` (array): Available token amounts from configuration

---

## Functions

### Public Event Handlers

#### Payment Type Selection
**Event:** `click` on `.payment-type-card`  
**Description:** Handles payment type selection (prepaid/postpaid). Blocks prepaid with informational message.

**Behavior:**
- Prepaid: Shows "Feature Unavailable" modal, prevents form display
- Postpaid: Shows payment form and initializes workflow

---

#### Check Customer Button
**Event:** `click` on `#checkCustomerBtn`  
**Description:** Validates customer meter number or ID via backend API

**Validations:**
- Input must not be empty
- Shows loading state during API call

**Response Handling:**
- Success: Displays customer details, shows appropriate payment options
- Error: Shows detailed error message with support link

---

#### Payment Method Selection
**Event:** `click` on `.payment-method-card` (delegated)  
**Description:** Handles payment method selection and updates summary

**Available Methods:**
- **QRIS**: QR code scanning (available for all)
- **Bank Transfer**: Disabled for postpaid
- **E-Wallet**: Disabled for postpaid

---

#### Form Submission
**Event:** `submit` on `#paymentForm`  
**Description:** Processes payment and sends to backend API

**Validations:**
- Customer data must be loaded
- Amount must be selected
- Payment method must be selected

---

### Private Functions

#### `initializePaymentTypeHandlers()`
**Parameters:** None  
**Returns:** void  
**Description:** Sets up vanilla JS event handlers for payment type selection with prepaid blocking logic

**Implementation Details:**
- Intercepts click events before jQuery handlers
- Uses `event.stopImmediatePropagation()` to prevent further processing
- Shows SweetAlert2 modal for unavailable features

---

#### `initializePaymentMethodRestrictions()`
**Parameters:** None  
**Returns:** void  
**Description:** Initializes MutationObserver to enforce payment method restrictions

**Behavior:**
- **Postpaid:** Only QR code available, others disabled with overlay
- **Prepaid:** All methods available (currently feature blocked)

---

#### `disablePaymentMethods(methods)`
**Parameters:**
- `methods` (string[]): Array of payment method types to disable

**Returns:** void  
**Description:** Disables specific payment methods with visual indicators

**Visual Changes:**
- Opacity: 50%
- Cursor: not-allowed
- Pointer events: disabled
- Overlay: "Tidak Tersedia" (Not Available)

---

#### `enableAllPaymentMethods()`
**Parameters:** None  
**Returns:** void  
**Description:** Removes all disabled states from payment method cards

---

#### `showQROnlyNotice()`
**Parameters:** None  
**Returns:** void  
**Description:** Displays informational alert about QR-only availability for postpaid

**Message:** "Untuk pembayaran tagihan listrik, saat ini hanya metode **Scan QR Code** yang tersedia."

---

#### `removeQROnlyNotice()`
**Parameters:** None  
**Returns:** void  
**Description:** Removes the QR-only notice from DOM

---

#### `checkCustomerData(customerInput)`
**Parameters:**
- `customerInput` (string): Customer meter number or ID

**Returns:** void  
**Description:** Makes AJAX request to verify customer existence in database

**API Endpoint:** `POST /api/check-customer`  
**Request Body:**
```json
{
  "customer_input": "string",
  "payment_type": "prepaid|postpaid"
}
```

**Success Response:**
```json
{
  "success": true,
  "data": {
    "meter_number": "string",
    "customer_id": "string",
    "connection": "prabayar|pascabayar",
    "tariff_code": "string",
    "power_va": number,
    "balance_kwh": number,
    "outstanding_balance": number,
    ...
  }
}
```

---

#### `showCustomerDetails(data)`
**Parameters:**
- `data` (object): Customer information object
  - `meter_number` (string): Electricity meter number
  - `customer_id` (string): Unique customer ID
  - `connection` (string): Connection type
  - `tariff_code` (string): Tariff classification
  - `power_va` (number): Power capacity in VA

**Returns:** void  
**Description:** Renders customer information in the UI

---

#### `showTokenOptions()`
**Parameters:** None  
**Returns:** void  
**Description:** Displays available token purchase amounts for prepaid customers

**Token Calculation:** Approximate kWh = Amount / 1500

**UI Elements:**
- Clickable cards with amount and kWh estimation
- Active state management
- Updates payment summary on selection

---

#### `showBillInfo(data)`
**Parameters:**
- `data` (object): Bill information
  - `outstanding_balance` (number): Current unpaid amount
  - `current_usage_kwh` (number): Monthly usage in kWh
  - `billing_cycle_start` (string): Period start date
  - `billing_cycle_end` (string): Period end date

**Returns:** void  
**Description:** Displays postpaid bill details including amount due and usage

---

#### `showPaymentMethods()`
**Parameters:** None  
**Returns:** void  
**Description:** Shows the payment method selection section

---

#### `updatePaymentSummary()`
**Parameters:** None  
**Returns:** void  
**Description:** Calculates and displays payment breakdown

**Calculation:**
- Base Amount: Selected token/bill amount
- Admin Fee: Rp 2,500 (constant)
- Total: Base Amount + Admin Fee

**Display Format:** IDR currency (Indonesian Rupiah)

---

#### `processPayment()`
**Parameters:** None  
**Returns:** void  
**Description:** Submits payment to backend API and handles response

**API Endpoint:** `POST /api/process-payment`  
**Request Body:**
```json
{
  "customer_id": "string",
  "payment_type": "prepaid|postpaid",
  "amount": number,
  "payment_method": "qr|bank|ewallet"
}
```

**Success Response (Prepaid):**
```json
{
  "success": true,
  "data": {
    "token": "string",
    "amount": number,
    "added_kwh": number,
    "new_balance": number
  }
}
```

**Success Response (Postpaid):**
```json
{
  "success": true,
  "data": {
    "paid_amount": number,
    "remaining_balance": number,
    "payment_date": "datetime"
  }
}
```

**Loading State:** Shows SweetAlert2 loading modal during processing

---

#### `resetForm()`
**Parameters:** None  
**Returns:** void  
**Description:** Resets all form data and returns UI to initial state

**Reset Actions:**
- Clears all selected values
- Hides all dynamic sections
- Removes active classes
- Clears input fields

---

## Payment Flow

### Prepaid (Currently Disabled)
1. User clicks "Token Listrik" card
2. System shows "Feature Unavailable" modal
3. Process stops, form not displayed

### Postpaid (Active)
1. User clicks "Tagihan Listrik" card
2. Payment form section appears
3. User enters meter number/customer ID
4. User clicks "Cek Data Pelanggan"
5. System validates customer via API
6. Bill information displays (amount, usage, period)
7. Payment methods show (only QR active, others disabled)
8. User selects QR code payment
9. Payment summary displays
10. User clicks "Proses Pembayaran"
11. System processes payment via API
12. Success/error message displays
13. Form resets on success

---

## Payment Method Restrictions

### Postpaid
- ✅ **QR Code (QRIS)**: Active
- ❌ **Bank Transfer**: Disabled (visual overlay)
- ❌ **E-Wallet**: Disabled (visual overlay)

**Notice:** "Untuk pembayaran tagihan listrik, saat ini hanya metode **Scan QR Code** yang tersedia."

### Prepaid (When Available)
- ✅ **QR Code (QRIS)**: Active
- ✅ **Bank Transfer**: Active
- ✅ **E-Wallet**: Active

---

## Dependencies

- **jQuery**: DOM manipulation and AJAX requests
- **SweetAlert2**: Modal dialogs and notifications
- **Bootstrap**: UI framework (cards, alerts, forms)
- **Boxicons**: Icons for UI elements

---

## Configuration

### Token Options
Set via `window.TokenOptions` variable passed from backend template:
```javascript
window.TokenOptions = "{{.TOKEN_OPTIONS}}";
```

Example values: [20000, 50000, 100000, 200000, 500000]

### Admin Fee
Hard-coded constant: **Rp 2,500**

---

## Error Handling

### Customer Not Found
- Shows detailed error message
- Provides troubleshooting steps
- Links to customer service section

### API Errors
- Generic error messages for user
- Console logging for debugging
- Button state restoration

### Validation Errors
- Input validation before API calls
- Warning messages for incomplete data
- Prevents submission with missing fields

---

## Future Enhancements

1. **Prepaid Feature**: Enable token purchase when ready
2. **Multiple Payment Methods**: Add bank transfer and e-wallet for postpaid
3. **Payment History**: Show transaction history
4. **Auto-fill**: Save customer data for repeat users
5. **Real-time Status**: WebSocket for payment status updates

---

## Browser Compatibility

- Modern browsers (Chrome, Firefox, Safari, Edge)
- IE11+ (with polyfills)
- Mobile responsive

---

## License

Copyright © 2025 PLTMH Lembang Palesan. All rights reserved.
