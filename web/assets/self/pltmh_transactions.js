// PLTMH Transaction Management JavaScript
let tableTransactions;

$(document).ready(function() {
    initializeTransactionsTable();
});

function initializeTransactionsTable() {
    tableTransactions = $('#tableTransactions').DataTable({
        processing: true,
        serverSide: false,
        ajax: {
            url: PLTMH_ENDPOINTS.transactions.table,
            type: 'POST',
            headers: {
                'X-CSRF-TOKEN': $('meta[name="csrf-token"]').attr('content')
            },
            dataSrc: 'data'
        },
        columns: [
            { data: 'transaction_id' },
            { 
                data: 'initiated_at',
                render: function(data) {
                    return new Date(data).toLocaleString('id-ID');
                }
            },
            { data: 'customer_id' },
            { data: 'meter_number' },
            { 
                data: 'payment_type',
                render: function(data) {
                    return data === 'prabayar' ? '<span class="badge bg-info">Prabayar</span>' : '<span class="badge bg-warning">Pascabayar</span>';
                }
            },
            { 
                data: 'payment_method',
                render: function(data) {
                    const methodMap = {
                        'qris': 'QRIS',
                        'bank_transfer': 'Transfer Bank',
                        'e_wallet': 'E-Wallet',
                        'manual': 'Manual'
                    };
                    return methodMap[data] || data;
                }
            },
            { 
                data: 'total_amount',
                render: function(data) {
                    return 'Rp ' + parseInt(data).toLocaleString('id-ID');
                }
            },
            {
                data: 'status',
                render: function(data) {
                    const statusMap = {
                        'pending': '<span class="badge bg-warning">Pending</span>',
                        'completed': '<span class="badge bg-success">Selesai</span>',
                        'failed': '<span class="badge bg-danger">Gagal</span>',
                        'expired': '<span class="badge bg-secondary">Kadaluarsa</span>'
                    };
                    return statusMap[data] || data;
                }
            },
            {
                data: null,
                orderable: false,
                render: function(data, type, row) {
                    return `
                        <button class="btn btn-sm btn-info" onclick="viewTransactionDetail('${row.transaction_id}')">
                            <i class="bx bx-show"></i> Detail
                        </button>
                    `;
                }
            }
        ],
        order: [[1, 'desc']],
        language: {
            url: '/assets/self/datatables/id.json'
        }
    });
}

// ===========================
// MANUAL TOP-UP (PREPAID)
// ===========================

function submitManualTopUp() {
    const form = document.getElementById('formManualTopUp');
    const formData = new FormData(form);
    const data = Object.fromEntries(formData);
    data.amount_rp = parseInt(data.amount_rp);
    data.added_kwh = parseFloat(data.added_kwh);

    $.ajax({
        url: PLTMH_ENDPOINTS.transactions.manualTopup,
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            if (response.success) {
                Swal.fire({
                    title: 'Top-Up Berhasil!',
                    html: `
                        <div class="text-start">
                            <p><strong>ID Pelanggan:</strong> ${response.data.customer_id}</p>
                            <p><strong>Token Listrik:</strong> <code>${response.data.token}</code></p>
                            <p><strong>Nominal:</strong> Rp ${parseInt(response.data.amount_paid).toLocaleString('id-ID')}</p>
                            <p><strong>kWh Ditambahkan:</strong> ${parseFloat(response.data.kwh_added).toFixed(2)} kWh</p>
                            <p><strong>Saldo Baru:</strong> ${parseFloat(response.data.balance_kwh).toFixed(2)} kWh</p>
                        </div>
                    `,
                    icon: 'success',
                    confirmButtonText: 'Tutup'
                });
                $('#modalManualTopUp').modal('hide');
                form.reset();
                tableTransactions.ajax.reload();
            } else {
                Swal.fire('Gagal!', response.message, 'error');
            }
        },
        error: function(xhr) {
            const error = xhr.responseJSON ? xhr.responseJSON.message : 'Terjadi kesalahan';
            Swal.fire('Kesalahan!', error, 'error');
        }
    });
}

// ===========================
// MANUAL PAYMENT (POSTPAID)
// ===========================

function submitManualPayment() {
    const form = document.getElementById('formManualPayment');
    const formData = new FormData(form);
    const data = Object.fromEntries(formData);
    data.amount_rp = parseInt(data.amount_rp);

    $.ajax({
        url: PLTMH_ENDPOINTS.transactions.manualPayment,
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            if (response.success) {
                Swal.fire({
                    title: 'Pembayaran Berhasil!',
                    html: `
                        <div class="text-start">
                            <p><strong>ID Pelanggan:</strong> ${response.data.customer_id}</p>
                            <p><strong>Nominal Dibayar:</strong> Rp ${parseInt(response.data.amount_paid).toLocaleString('id-ID')}</p>
                            <p><strong>Sisa Tagihan:</strong> Rp ${parseInt(response.data.outstanding_balance).toLocaleString('id-ID')}</p>
                            <p><strong>Status:</strong> ${response.data.is_disconnected ? '<span class="badge bg-danger">Terputus</span>' : '<span class="badge bg-success">Aktif</span>'}</p>
                        </div>
                    `,
                    icon: 'success',
                    confirmButtonText: 'OK'
                });
                $('#modalManualPayment').modal('hide');
                form.reset();
                tableTransactions.ajax.reload();
            } else {
                Swal.fire('Gagal!', response.message, 'error');
            }
        },
        error: function(xhr) {
            const error = xhr.responseJSON ? xhr.responseJSON.message : 'Terjadi kesalahan';
            Swal.fire('Kesalahan!', error, 'error');
        }
    });
}

// ===========================
// ADD USAGE (POSTPAID)
// ===========================

function submitAddUsage() {
    const form = document.getElementById('formAddUsage');
    const formData = new FormData(form);
    const data = Object.fromEntries(formData);
    data.usage_kwh = parseFloat(data.usage_kwh);
    data.rate_per_kwh = parseFloat(data.rate_per_kwh);

    $.ajax({
        url: PLTMH_ENDPOINTS.transactions.addUsage,
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            if (response.success) {
                Swal.fire({
                    title: 'Pemakaian Ditambahkan!',
                    html: `
                        <div class="text-start">
                            <p><strong>ID Pelanggan:</strong> ${response.data.customer_id}</p>
                            <p><strong>Pemakaian Ditambahkan:</strong> ${parseFloat(response.data.usage_kwh).toFixed(2)} kWh</p>
                            <p><strong>Total Pemakaian Bulan Ini:</strong> ${parseFloat(response.data.current_usage_kwh).toFixed(2)} kWh</p>
                            <p><strong>Tagihan Sekarang:</strong> Rp ${parseInt(response.data.outstanding_balance).toLocaleString('id-ID')}</p>
                        </div>
                    `,
                    icon: 'success',
                    confirmButtonText: 'OK'
                });
                $('#modalAddUsage').modal('hide');
                form.reset();
                tableTransactions.ajax.reload();
            } else {
                Swal.fire('Gagal!', response.message, 'error');
            }
        },
        error: function(xhr) {
            const error = xhr.responseJSON ? xhr.responseJSON.message : 'Terjadi kesalahan';
            Swal.fire('Kesalahan!', error, 'error');
        }
    });
}

// ===========================
// VIEW TRANSACTION DETAIL
// ===========================

function viewTransactionDetail(transactionId) {
    // Get transaction data from table
    const rowData = tableTransactions.rows().data().toArray().find(row => row.transaction_id === transactionId);
    
    if (!rowData) {
        Swal.fire('Kesalahan!', 'Data tidak ditemukan', 'error');
        return;
    }

    let html = `
        <div class="row">
            <div class="col-md-6">
                <h6>Informasi Transaksi</h6>
                <table class="table table-borderless table-sm">
                    <tr><td><strong>ID Transaksi:</strong></td><td>${rowData.transaction_id}</td></tr>
                    <tr><td><strong>Tanggal:</strong></td><td>${new Date(rowData.initiated_at).toLocaleString('id-ID')}</td></tr>
                    <tr><td><strong>Status:</strong></td><td>${getStatusBadge(rowData.status)}</td></tr>
                    <tr><td><strong>Tipe Pembayaran:</strong></td><td>${rowData.payment_type === 'prabayar' ? 'Prabayar' : 'Pascabayar'}</td></tr>
                    <tr><td><strong>Metode:</strong></td><td>${getPaymentMethod(rowData.payment_method)}</td></tr>
                </table>
            </div>
            <div class="col-md-6">
                <h6>Informasi Pelanggan</h6>
                <table class="table table-borderless table-sm">
                    <tr><td><strong>ID Pelanggan:</strong></td><td>${rowData.customer_id}</td></tr>
                    <tr><td><strong>No. Meter:</strong></td><td>${rowData.meter_number}</td></tr>
                </table>
                <h6 class="mt-3">Rincian Biaya</h6>
                <table class="table table-borderless table-sm">
                    <tr><td><strong>Nominal:</strong></td><td>Rp ${parseInt(rowData.amount).toLocaleString('id-ID')}</td></tr>
                    <tr><td><strong>Biaya Admin:</strong></td><td>Rp ${parseInt(rowData.admin_fee).toLocaleString('id-ID')}</td></tr>
                    <tr><td><strong>Total:</strong></td><td><strong>Rp ${parseInt(rowData.total_amount).toLocaleString('id-ID')}</strong></td></tr>
                </table>
            </div>
        </div>
    `;

    if (rowData.payment_type === 'prabayar' && rowData.token) {
        html += `
            <div class="mt-3">
                <h6>Token Listrik</h6>
                <div class="alert alert-info">
                    <p class="mb-0"><strong>Token:</strong> <code>${rowData.token}</code></p>
                    <p class="mb-0"><strong>kWh Ditambahkan:</strong> ${parseFloat(rowData.added_kwh).toFixed(2)} kWh</p>
                </div>
            </div>
        `;
    }

    if (rowData.qr_code_url) {
        html += `
            <div class="mt-3">
                <h6>QR Code Pembayaran</h6>
                <img src="${rowData.qr_code_url}" alt="QR Code" class="img-fluid" style="max-width: 200px;">
            </div>
        `;
    }

    if (rowData.completed_at) {
        html += `
            <div class="mt-3">
                <p><strong>Selesai pada:</strong> ${new Date(rowData.completed_at).toLocaleString('id-ID')}</p>
            </div>
        `;
    }

    if (rowData.expired_at) {
        html += `
            <div class="mt-3">
                <p><strong>Kadaluarsa pada:</strong> ${new Date(rowData.expired_at).toLocaleString('id-ID')}</p>
            </div>
        `;
    }

    $('#transactionDetailBody').html(html);
    $('#modalTransactionDetail').modal('show');
}

function getStatusBadge(status) {
    const statusMap = {
        'pending': '<span class="badge bg-warning">Pending</span>',
        'completed': '<span class="badge bg-success">Selesai</span>',
        'failed': '<span class="badge bg-danger">Gagal</span>',
        'expired': '<span class="badge bg-secondary">Kadaluarsa</span>'
    };
    return statusMap[status] || status;
}

function getPaymentMethod(method) {
    const methodMap = {
        'qris': 'QRIS',
        'bank_transfer': 'Transfer Bank',
        'e_wallet': 'E-Wallet',
        'manual': 'Manual'
    };
    return methodMap[method] || method;
}
