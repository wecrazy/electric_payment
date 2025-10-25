$(document).ready(function () {
    let selectedPaymentType = '';
    let selectedAmount = 0;
    let selectedPaymentMethod = '';
    let customerData = null;

    // Token options from config  
    const tokenOptions = TokenOptions;
    console.log('Token Options:', tokenOptions);

    // Payment type selection
    $('.payment-type-card').on('click', function () {
        selectedPaymentType = $(this).data('type');

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

        const adminFee = 2500; // Admin fee
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
                if (data.success) {
                    let successMessage = `<p>Pembayaran ${selectedPaymentType === 'prepaid' ? 'token' : 'tagihan'} listrik Anda telah berhasil diproses.</p>`;
                    
                    if (selectedPaymentType === 'prepaid') {
                        successMessage += `<p><strong>Token Anda: ${data.data.token || 'Token akan dikirimkan'}</strong></p>`;
                        successMessage += '<p>Token akan dikirimkan ke WhatsApp Anda dalam 5 menit.</p>';
                    } else {
                        successMessage += `<p>Sisa tagihan: ${new Intl.NumberFormat('id-ID', {
                            style: 'currency',
                            currency: 'IDR',
                            minimumFractionDigits: 0
                        }).format(data.data.remaining_balance || 0)}</p>`;
                    }

                    Swal.fire({
                        icon: 'success',
                        title: 'Pembayaran Berhasil',
                        html: successMessage,
                        confirmButtonText: 'OK'
                    }).then(() => {
                        resetForm();
                    });
                } else {
                    Swal.fire({
                        icon: 'error',
                        title: 'Pembayaran Gagal',
                        text: data.message || 'Terjadi kesalahan saat memproses pembayaran.',
                        confirmButtonText: 'OK'
                    });
                }
            },
            error: function (xhr, status, error) {
                console.error('Payment Error:', error);
                Swal.fire({
                    icon: 'error',
                    title: 'Terjadi Kesalahan',
                    text: 'Tidak dapat memproses pembayaran. Silakan coba lagi.',
                    confirmButtonText: 'OK'
                });
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
});
