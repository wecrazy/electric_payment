$(document).ready(function () {
    let selectedPaymentType = '';
    let selectedAmount = 0;
    let selectedPaymentMethod = '';
    let customerData = null;
    let prepaidBlocked = false;

    // Token options from config  
    const tokenOptions = TokenOptions;
    console.log('Token Options:', tokenOptions);

    /**
     * Initialize payment type selection handlers
     * Sets up vanilla JS handlers for prepaid blocking and postpaid restrictions
     * This runs before jQuery handlers to intercept prepaid clicks
     */
    initializePaymentTypeHandlers();

    /**
     * Initialize payment method restrictions observer
     * Sets up MutationObserver to handle payment method availability based on payment type
     */
    initializePaymentMethodRestrictions();

    /**
     * Handles prepaid blocking and postpaid selection
     * Prevents prepaid option from proceeding and shows unavailable message
     * 
     * @private
     */
    function initializePaymentTypeHandlers() {
        document.querySelectorAll('.payment-type-card').forEach(card => {
            card.addEventListener('click', function(event) {
                const type = this.dataset.type;
                if (type === 'prepaid') {
                    event.preventDefault();
                    event.stopPropagation();
                    event.stopImmediatePropagation();
                    prepaidBlocked = true;
                    window.prepaidBlocked = true;
                    
                    Swal.fire({
                        title: 'Fitur Belum Tersedia',
                        text: 'Pembelian token listrik secara online sedang dalam pengembangan. Fitur ini akan segera hadir untuk memudahkan Anda 😃',
                        icon: 'info',
                        confirmButtonText: 'Mengerti'
                    });
                    return false;
                }
                prepaidBlocked = false;
                window.prepaidBlocked = false;
            });
        });
    }

    /**
     * Initializes MutationObserver to restrict payment methods based on payment type
     * For postpaid: Only QR code is available, bank and e-wallet are disabled
     * For prepaid: All methods would be available (currently blocked)
     * 
     * @private
     */
    function initializePaymentMethodRestrictions() {
        const paymentMethodSection = document.getElementById('paymentMethodSection');
        if (!paymentMethodSection) return;
        
        const methodObserver = new MutationObserver(function(mutations) {
            mutations.forEach(function(mutation) {
                if (mutation.type === 'attributes' && mutation.attributeName === 'class') {
                    if (!paymentMethodSection.classList.contains('hidden') && selectedPaymentType === 'postpaid') {
                        disablePaymentMethods(['bank', 'ewallet']);
                        showQROnlyNotice();
                    } else {
                        enableAllPaymentMethods();
                        removeQROnlyNotice();
                    }
                }
            });
        });
        methodObserver.observe(paymentMethodSection, { attributes: true });
    }

    /**
     * Disables specific payment methods by adding visual indicators and preventing interaction
     * 
     * @param {string[]} methods - Array of payment method types to disable (e.g., ['bank', 'ewallet'])
     * @private
     */
    function disablePaymentMethods(methods) {
        methods.forEach(method => {
            const selector = `.payment-method-card[data-method="${method}"]`;
            document.querySelectorAll(selector).forEach(card => {
                card.style.opacity = '0.5';
                card.style.cursor = 'not-allowed';
                card.style.pointerEvents = 'none';
                
                // Add disabled overlay with blur effect
                if (!card.querySelector('.disabled-overlay')) {
                    const overlay = document.createElement('div');
                    overlay.className = 'disabled-overlay';
                    overlay.style.cssText = 'position: absolute; top: 0; left: 0; right: 0; bottom: 0; background: rgba(255,255,255,0.85); backdrop-filter: blur(3px); -webkit-backdrop-filter: blur(3px); border-radius: 0.375rem; display: flex; align-items: center; justify-content: center;';
                    overlay.innerHTML = '<small class="text-dark fw-bold" style="font-size: 0.85rem; text-shadow: 0 1px 2px rgba(255,255,255,0.8);">Tidak Tersedia</small>';
                    card.style.position = 'relative';
                    card.appendChild(overlay);
                }
            });
        });
    }

    /**
     * Enables all payment methods by removing disabled state and overlays
     * 
     * @private
     */
    function enableAllPaymentMethods() {
        document.querySelectorAll('.payment-method-card').forEach(card => {
            card.style.opacity = '';
            card.style.cursor = '';
            card.style.pointerEvents = '';
            const overlay = card.querySelector('.disabled-overlay');
            if (overlay) overlay.remove();
        });
    }

    /**
     * Shows informational notice that only QR code payment is available
     * 
     * @private
     */
    function showQROnlyNotice() {
        const paymentMethodSection = document.getElementById('paymentMethodSection');
        if (!document.getElementById('qr-only-notice')) {
            const notice = document.createElement('div');
            notice.id = 'qr-only-notice';
            notice.className = 'alert alert-info mt-3';
            notice.innerHTML = '<i class="bx bx-info-circle me-2"></i>Untuk pembayaran tagihan listrik, saat ini hanya metode <strong>Scan QR Code</strong> yang tersedia.';
            paymentMethodSection.appendChild(notice);
        }
    }

    /**
     * Removes the QR-only notice from the payment method section
     * 
     * @private
     */
    function removeQROnlyNotice() {
        const notice = document.getElementById('qr-only-notice');
        if (notice) notice.remove();
    }

    // Payment type selection
    $('.payment-type-card').on('click', function () {
        // Check if prepaid is blocked
        if (window.prepaidBlocked || prepaidBlocked) {
            return;
        }

        selectedPaymentType = $(this).data('type');

        // Block prepaid from proceeding
        if (selectedPaymentType === 'prepaid') {
            return;
        }

        // Remove active class from all cards
        $('.payment-option').removeClass('border-primary bg-light').addClass('border');

        // Add active class to selected card
        $(this).find('.payment-option').removeClass('border').addClass('border-primary bg-light');

        // Show payment form
        $('#paymentFormSection').removeClass('hidden');

        // Update form title
        const title = selectedPaymentType === 'prepaid' ? 'Pembelian Token Listrik (Prabayar)' : 'Pembayaran Tagihan Listrik (Pascabayar)';
        $('#paymentFormTitle').text(title);

        // Reset form
        resetForm();
    });

    // Check customer button
    $('#checkCustomerBtn').on('click', function () {
        const customerInput = $('#customerInput').val().trim();

        if (!customerInput) {
            Swal.fire({
                icon: 'warning',
                title: 'Input Kosong',
                text: 'Silakan masukkan No. Meter atau ID Pelanggan terlebih dahulu.',
                confirmButtonText: 'OK'
            });
            return;
        }

        // Show loading
        $(this).html('<i class="bx bx-loader-alt bx-spin me-2"></i>Mengecek...').prop('disabled', true);

        // Check customer data
        checkCustomerData(customerInput);
    });

    // Payment method selection
    $(document).on('click', '.payment-method-card', function () {
        selectedPaymentMethod = $(this).data('method');

        // Remove active class from all cards
        $('.payment-method-card').removeClass('border-primary bg-light').addClass('border');

        // Add active class to selected card
        $(this).removeClass('border').addClass('border-primary bg-light');

        updatePaymentSummary();
    });

    // Form submission
    $('#paymentForm').on('submit', function (e) {
        e.preventDefault();
        processPayment();
    });

    function checkCustomerData(customerInput) {
        $.ajax({
            url: '/api/check-customer',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({
                customer_input: customerInput,
                payment_type: selectedPaymentType
            }),
            success: function (data) {
                $('#checkCustomerBtn')
                    .html('<i class="bx bx-search me-2"></i>Cek Data Pelanggan')
                    .prop('disabled', false);

                if (data.success) {
                    customerData = data.data;
                    showCustomerDetails(data.data);

                    if (selectedPaymentType === 'prepaid') {
                        showTokenOptions();
                    } else {
                        showBillInfo(data.data);
                    }

                    showPaymentMethods();
                } else {
                    // Check if customer is disconnected
                    if (data.is_disconnected) {
                        Swal.fire({
                            icon: 'error',
                            title: 'Sambungan Terputus',
                            html: `
                                <p>${data.message}</p>
                                <div class="mt-3">
                                    <strong>Hubungi Customer Service:</strong><br>
                                    <a href="https://wa.me/${data.support_contact.replace(/\+/g, '')}" 
                                       class="btn btn-success mt-2" target="_blank">
                                        <i class='bx bxl-whatsapp'></i> ${data.support_contact}
                                    </a>
                                </div>
                            `,
                            confirmButtonText: 'OK',
                            confirmButtonColor: '#d33',
                            footer: '<small>Layanan akan diaktifkan kembali setelah menghubungi CS</small>'
                        });
                    } else if (data.no_bill) {
                        // Customer has no outstanding bill
                        Swal.fire({
                            icon: 'info',
                            title: 'Tidak Ada Tagihan',
                            html: `
                                <p>${data.message}</p>
                                <div class="mt-3">
                                    <i class="bx bx-check-circle" style="font-size: 48px; color: #28a745;"></i>
                                    <p class="mt-2"><strong>Status Pembayaran: Lunas</strong></p>
                                </div>
                            `,
                            confirmButtonText: 'OK',
                            confirmButtonColor: '#28a745'
                        });
                    } else {
                        Swal.fire({
                            icon: 'error',
                            title: 'Data Tidak Ditemukan',
                            html: `
                                <p>Nomor meter atau ID pelanggan <strong>${customerInput}</strong> tidak terdaftar dalam sistem kami.</p>
                                <br>
                                <p>Silakan:</p>
                                <ul style="text-align: left; margin: 0 auto; display: inline-block;">
                                    <li>Periksa kembali nomor yang dimasukkan</li>
                                    <li>Hubungi customer service jika masih bermasalah</li>
                                </ul>
                            `,
                            confirmButtonText: 'OK',
                            footer: '<a href="#landingContact">Hubungi Customer Service</a>'
                        });
                    }
                }
            },
            error: function (xhr, status, error) {
                console.error('Error:', error);
                $('#checkCustomerBtn')
                    .html('<i class="bx bx-search me-2"></i>Cek Data Pelanggan')
                    .prop('disabled', false);

                Swal.fire({
                    icon: 'error',
                    title: 'Terjadi Kesalahan',
                    text: 'Tidak dapat mengecek data pelanggan. Silakan coba lagi.',
                    confirmButtonText: 'OK'
                });
            }
        });
    }

    function showCustomerDetails(data) {
        const customerInfoHtml = `
                    <div class="row">
                        <div class="col-md-6">
                            <strong>No. Meter:</strong> ${data.meter_number}<br>
                            <strong>ID Pelanggan:</strong> ${data.customer_id}<br>
                            <strong>Jenis Sambungan:</strong> ${data.connection === 'prabayar' ? 'Prabayar (Token)' : 'Pascabayar (Tagihan)'}
                        </div>
                        <div class="col-md-6">
                            <strong>Tarif:</strong> ${data.tariff_code}<br>
                            <strong>Daya:</strong> ${data.power_va} VA<br>
                            <strong>Status:</strong> <span class="badge bg-success">Aktif</span>
                        </div>
                    </div>
                `;
        
        $('#customerInfo').html(customerInfoHtml);
        $('#customerDetails').removeClass('hidden');
    }

    function showTokenOptions() {
        const $tokenOptionsContainer = $('#tokenOptions');
        $tokenOptionsContainer.empty();

        tokenOptions.forEach(amount => {
            const formatAmount = new Intl.NumberFormat('id-ID', {
                style: 'currency',
                currency: 'IDR',
                minimumFractionDigits: 0
            }).format(amount);

            const tokenOptionHtml = `
                        <div class="col-md-4 col-sm-6 mb-3">
                            <div class="card border cursor-pointer token-option-card" data-amount="${amount}">
                                <div class="card-body text-center">
                                    <h5 class="text-primary mb-2">${formatAmount}</h5>
                                    <small class="text-muted">≈ ${Math.round(amount / 1500)} kWh</small>
                                </div>
                            </div>
                        </div>
                    `;

            $tokenOptionsContainer.append(tokenOptionHtml);
        });

        // Add event listeners for token options using event delegation
        $tokenOptionsContainer.off('click', '.token-option-card').on('click', '.token-option-card', function () {
            selectedAmount = parseInt($(this).data('amount'));

            // Remove active class from all cards
            $('.token-option-card').removeClass('border-primary bg-light').addClass('border');

            // Add active class to selected card
            $(this).removeClass('border').addClass('border-primary bg-light');

            updatePaymentSummary();
        });

        $('#tokenAmountSection').removeClass('hidden');
    }

    function showBillInfo(data) {
        const currentBill = data.outstanding_balance || 0;
        const usage = data.current_usage_kwh || 0;

        const formatCurrency = new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            minimumFractionDigits: 0
        }).format(currentBill);

        const billDetailsHtml = `
                    <div class="row">
                        <div class="col-md-6">
                            <strong>Tagihan Bulan Ini:</strong><br>
                            <h4 class="text-warning">${formatCurrency}</h4>
                            <small class="text-muted">Pemakaian: ${usage} kWh</small>
                        </div>
                        <div class="col-md-6">
                            <strong>Periode Tagihan:</strong><br>
                            ${data.billing_cycle_start || 'N/A'} - ${data.billing_cycle_end || 'N/A'}<br>
                            <small class="text-muted">Jatuh tempo: ${data.billing_cycle_end || 'N/A'}</small>
                        </div>
                    </div>
                `;

        $('#billDetails').html(billDetailsHtml);
        selectedAmount = currentBill;
        $('#billInfoSection').removeClass('hidden');
        updatePaymentSummary();
    }

    function showPaymentMethods() {
        $('#paymentMethodSection').removeClass('hidden');
    }

    function updatePaymentSummary() {
        if (!selectedAmount || !selectedPaymentMethod) return;

        let adminFee = window.AdminFee; // Admin fee
        const total = selectedAmount + adminFee;

        const formatCurrency = new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            minimumFractionDigits: 0
        }).format;

        let paymentMethodText = '';
        switch (selectedPaymentMethod) {
            case 'qr': paymentMethodText = 'QRIS (Scan QR Code)'; break;
            case 'bank': paymentMethodText = 'Transfer Bank'; break;
            case 'ewallet': paymentMethodText = 'E-Wallet'; break;
        }

        const itemText = selectedPaymentType === 'prepaid' ? 'Token Listrik' : 'Tagihan Listrik';

        const summaryHtml = `
                    <div class="row">
                        <div class="col-8"><strong>${itemText}:</strong></div>
                        <div class="col-4 text-end">${formatCurrency(selectedAmount)}</div>
                    </div>
                    <div class="row">
                        <div class="col-8">Biaya Admin:</div>
                        <div class="col-4 text-end">${formatCurrency(adminFee)}</div>
                    </div>
                    <hr>
                    <div class="row">
                        <div class="col-8"><strong>Total Pembayaran:</strong></div>
                        <div class="col-4 text-end"><strong class="text-primary">${formatCurrency(total)}</strong></div>
                    </div>
                    <div class="row mt-2">
                        <div class="col-12">
                            <small class="text-muted">Metode: ${paymentMethodText}</small>
                        </div>
                    </div>
                `;

        $('#paymentSummary').html(summaryHtml);
        $('#paymentSummarySection').removeClass('hidden');
        $('#processPaymentBtn').removeClass('hidden');
    }

    function processPayment() {
        if (!customerData || !selectedAmount || !selectedPaymentMethod) {
            Swal.fire({
                icon: 'warning',
                title: 'Data Tidak Lengkap',
                text: 'Silakan lengkapi semua data yang diperlukan.',
                confirmButtonText: 'OK'
            });
            return;
        }

        console.log('Processing payment with data:', {
            customer_id: customerData.customer_id,
            payment_type: selectedPaymentType,
            amount: selectedAmount,
            payment_method: selectedPaymentMethod
        });

        Swal.fire({
            icon: 'info',
            title: 'Memproses Pembayaran',
            html: 'Silakan tunggu, pembayaran Anda sedang diproses...',
            allowOutsideClick: false,
            didOpen: () => {
                Swal.showLoading();
            }
        });

        // Make API call to process payment
        $.ajax({
            url: '/api/process-payment',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({
                customer_id: customerData.customer_id,
                payment_type: selectedPaymentType,
                amount: selectedAmount,
                payment_method: selectedPaymentMethod
            }),
            success: function (data) {
                console.log('Payment response:', data);
                
                if (data.success) {
                    // Show QR code for payment
                    showQRCodePayment(data.data);
                } else {
                    // Show detailed error message from backend
                    let errorHtml = data.message || 'Terjadi kesalahan saat memproses pembayaran.';
                    
                    if (data.error) {
                        errorHtml += `<br><br><small class="text-muted">Detail: ${data.error}</small>`;
                    }
                    
                    Swal.fire({
                        icon: 'error',
                        title: 'Pembayaran Gagal',
                        html: errorHtml,
                        footer: '<small>Silakan periksa data dan coba lagi</small>',
                        confirmButtonText: 'OK'
                    });
                }
            },
            error: function (xhr, status, error) {
                console.error('Payment Error:', error);
                console.error('Response:', xhr.responseText);
                console.error('Status:', xhr.status);
                
                let errorMessage = 'Tidak dapat memproses pembayaran. Silakan coba lagi.';
                let errorDetails = '';
                let isDisconnected = false;
                let noBill = false;
                let supportContact = '';
                
                // Try to parse error response
                try {
                    const response = JSON.parse(xhr.responseText);
                    if (response.message) {
                        errorMessage = response.message;
                    }
                    if (response.error) {
                        errorDetails = `<br><small class="text-muted">Detail: ${response.error}</small>`;
                    }
                    if (response.is_disconnected) {
                        isDisconnected = true;
                        supportContact = response.support_contact || '';
                    }
                    if (response.no_bill) {
                        noBill = true;
                    }
                } catch (e) {
                    // If can't parse JSON, show status code
                    if (xhr.status) {
                        errorDetails = `<br><small class="text-muted">Error Code: ${xhr.status}</small>`;
                    }
                }
                
                // Show special message for no bill
                if (noBill) {
                    Swal.fire({
                        icon: 'info',
                        title: 'Tidak Ada Tagihan',
                        html: `
                            <p>${errorMessage}</p>
                            <div class="mt-3">
                                <i class="bx bx-check-circle" style="font-size: 48px; color: #28a745;"></i>
                                <p class="mt-2"><strong>Status Pembayaran: Lunas</strong></p>
                                <p class="text-muted">Terima kasih atas pembayaran Anda yang tepat waktu!</p>
                            </div>
                        `,
                        confirmButtonText: 'OK',
                        confirmButtonColor: '#28a745'
                    });
                }
                // Show special message for disconnected customers
                else if (isDisconnected && supportContact) {
                    Swal.fire({
                        icon: 'error',
                        title: 'Sambungan Terputus',
                        html: `
                            <p>${errorMessage}</p>
                            <div class="mt-3">
                                <strong>Hubungi Customer Service:</strong><br>
                                <a href="https://wa.me/${supportContact.replace(/\+/g, '')}" 
                                   class="btn btn-success mt-2" target="_blank">
                                    <i class='bx bxl-whatsapp'></i> ${supportContact}
                                </a>
                            </div>
                        `,
                        footer: '<small>Layanan akan diaktifkan kembali setelah menghubungi CS</small>',
                        confirmButtonText: 'OK',
                        confirmButtonColor: '#d33'
                    });
                } else {
                    Swal.fire({
                        icon: 'error',
                        title: 'Terjadi Kesalahan',
                        html: errorMessage + errorDetails,
                        footer: '<small>Jika masalah berlanjut, silakan hubungi customer service</small>',
                        confirmButtonText: 'OK'
                    });
                }
            }
        });
    }

    function resetForm() {
        // Reset all form data
        selectedAmount = 0;
        selectedPaymentMethod = '';
        customerData = null;

        // Hide sections
        $('#customerDetails').addClass('hidden');
        $('#tokenAmountSection').addClass('hidden');
        $('#billInfoSection').addClass('hidden');
        $('#paymentMethodSection').addClass('hidden');
        $('#paymentSummarySection').addClass('hidden');
        $('#processPaymentBtn').addClass('hidden');

        // Reset input
        $('#customerInput').val('');

        // Remove active classes
        $('.border-primary').removeClass('border-primary bg-light').addClass('border');
    }

    
/**
 * Shows QR code for payment and starts polling for payment status
 * 
 * @param {Object} paymentData - Payment data from backend
 * @private
 */
function showQRCodePayment(paymentData) {
    const expiresAt = new Date(paymentData.expires_at);
    const formatCurrency = new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0
    }).format;

    let qrModalHtml = `
        <div class="text-center">
            <h5 class="mb-3">Scan QR Code untuk Membayar</h5>
            <p class="text-muted">ID Transaksi: <strong>${paymentData.transaction_id}</strong></p>
            <div class="qr-code-container mb-3" style="background: white; padding: 20px; border-radius: 10px; display: inline-block;">
                <img src="${paymentData.qr_code_url}" alt="QR Code Pembayaran" style="max-width: 300px; width: 100%;">
            </div>
            <div class="payment-info mb-3">
                <p><strong>Total Pembayaran:</strong> ${formatCurrency(paymentData.total_amount)}</p>
                <p class="text-warning"><i class='bx bx-time'></i> Berlaku hingga: ${expiresAt.toLocaleString('id-ID')}</p>
            </div>
            <div class="alert alert-info">
                <i class='bx bx-info-circle me-2'></i>
                Halaman ini akan otomatis diperbarui saat pembayaran berhasil.
            </div>
            <div id="payment-status-indicator" class="mt-3">
                <div class="spinner-border text-primary" role="status">
                    <span class="visually-hidden">Menunggu pembayaran...</span>
                </div>
                <p class="mt-2">Menunggu pembayaran...</p>
            </div>
        </div>
    `;

    Swal.fire({
        html: qrModalHtml,
        width: 600,
        showCancelButton: true,
        cancelButtonText: 'Batal',
        showConfirmButton: false,
        allowOutsideClick: false,
        didOpen: () => {
            // Start polling for payment status
            startPaymentStatusPolling(paymentData.transaction_id);
        },
        willClose: () => {
            // Stop polling when modal closes
            stopPaymentStatusPolling();
        }
    });
}

let paymentPollingInterval = null;

/**
 * Starts polling the payment status
 * 
 * @param {string} transactionId - Transaction ID to check
 * @private
 */
function startPaymentStatusPolling(transactionId) {
    // Poll every 3 seconds
    paymentPollingInterval = setInterval(() => {
        checkPaymentStatus(transactionId);
    }, 3000);
}

/**
 * Stops the payment status polling
 * 
 * @private
 */
function stopPaymentStatusPolling() {
    if (paymentPollingInterval) {
        clearInterval(paymentPollingInterval);
        paymentPollingInterval = null;
    }
}

/**
 * Checks the payment status via API
 * 
 * @param {string} transactionId - Transaction ID to check
 * @private
 */
function checkPaymentStatus(transactionId) {
    $.ajax({
        url: `/api/payment-status/${transactionId}`,
        method: 'GET',
        success: function (data) {
            if (data.success && data.data.status === 'completed') {
                stopPaymentStatusPolling();
                Swal.close();
                
                Swal.fire({
                    icon: 'success',
                    title: 'Pembayaran Berhasil!',
                    html: `<p>Pembayaran Anda telah berhasil diproses.</p>
                           <p class="text-muted">ID Transaksi: ${transactionId}</p>`,
                    confirmButtonText: 'OK'
                }).then(() => {
                    resetForm();
                    // Optionally reload to show updated data
                    location.reload();
                });
            } else if (data.success && data.data.status === 'expired') {
                stopPaymentStatusPolling();
                Swal.close();
                
                Swal.fire({
                    icon: 'warning',
                    title: 'Transaksi Kadaluarsa',
                    text: 'Transaksi telah kadaluarsa. Silakan buat transaksi baru.',
                    confirmButtonText: 'OK'
                }).then(() => {
                    resetForm();
                });
            }
        },
        error: function (xhr, status, error) {
            console.error('Status check error:', error);
        }
    });
}

});
