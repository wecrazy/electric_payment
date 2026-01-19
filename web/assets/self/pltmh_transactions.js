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
                        'bank': 'Transfer Bank',
                        'ewallet': 'E-Wallet',
                        'manual': 'Manual'
                    };
                    return methodMap[data] || data.toUpperCase();
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
                        <div class="btn-group" role="group">
                            <button class="btn btn-sm btn-info" onclick="viewTransactionDetail('${row.transaction_id}')" title="Lihat Detail">
                                <i class="bx bx-show"></i>
                            </button>
                            <button class="btn btn-sm btn-success" onclick="exportTransaction('${row.transaction_id}', 'excel')" title="Export Excel">
                                <i class="fas fa-file-excel"></i>
                            </button>
                            <button class="btn btn-sm btn-danger" onclick="exportTransaction('${row.transaction_id}', 'pdf')" title="Export PDF">
                                <i class="fas fa-file-pdf"></i>
                            </button>
                        </div>
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
                    <tr><td><strong>No. Meter:</strong></td><td>${rowData.meter_number || '-'}</td></tr>
                </table>
                <h6 class="mt-3">Rincian Biaya</h6>
                <table class="table table-borderless table-sm">
                    <tr><td><strong>Nominal:</strong></td><td>Rp ${parseInt(rowData.amount || 0).toLocaleString('id-ID')}</td></tr>
                    <tr><td><strong>Biaya Admin:</strong></td><td>Rp ${parseInt(rowData.admin_fee || 0).toLocaleString('id-ID')}</td></tr>
                    <tr><td><strong>Total:</strong></td><td><strong>Rp ${parseInt(rowData.total_amount || 0).toLocaleString('id-ID')}</strong></td></tr>
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
                    <p class="mb-0"><strong>kWh Ditambahkan:</strong> ${parseFloat(rowData.added_kwh || 0).toFixed(2)} kWh</p>
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
        'bank': 'Transfer Bank',
        'ewallet': 'E-Wallet',
        'manual': 'Manual'
    };
    return methodMap[method] || method.toUpperCase();
}

// ===========================
// EXPORT FUNCTIONS
// ===========================

/**
 * Export a single transaction
 * @param {string} transactionId - The transaction ID to export
 * @param {string} format - 'excel' or 'pdf'
 */
function exportTransaction(transactionId, format) {
    // Get transaction data from table
    const rowData = tableTransactions.rows().data().toArray().find(row => row.transaction_id === transactionId);
    
    if (!rowData) {
        Swal.fire('Kesalahan!', 'Data tidak ditemukan', 'error');
        return;
    }

    if (format === 'excel') {
        exportTransactionToExcel(rowData);
    } else if (format === 'pdf') {
        exportTransactionToPDF(rowData);
    }
}

/**
 * Export a single transaction to Excel
 */
function exportTransactionToExcel(rowData) {
    // Prepare data for export
    const exportData = [
        ['DETAIL TRANSAKSI', ''],
        ['', ''],
        ['ID Transaksi', rowData.transaction_id],
        ['Tanggal Dibuat', new Date(rowData.initiated_at).toLocaleString('id-ID')],
        ['Status', rowData.status.toUpperCase()],
        ['', ''],
        ['INFORMASI PELANGGAN', ''],
        ['ID Pelanggan', rowData.customer_id],
        ['No. Meter', rowData.meter_number || '-'],
        ['Tipe Pembayaran', rowData.payment_type === 'prabayar' ? 'Prabayar' : 'Pascabayar'],
        ['', ''],
        ['RINCIAN BIAYA', ''],
        ['Nominal', 'Rp ' + parseInt(rowData.amount || 0).toLocaleString('id-ID')],
        ['Biaya Admin', 'Rp ' + parseInt(rowData.admin_fee || 0).toLocaleString('id-ID')],
        ['Total', 'Rp ' + parseInt(rowData.total_amount || 0).toLocaleString('id-ID')],
        ['Metode Pembayaran', getPaymentMethod(rowData.payment_method)]
    ];

    // Add token info if prepaid
    if (rowData.payment_type === 'prabayar' && rowData.token) {
        exportData.push(['', '']);
        exportData.push(['TOKEN LISTRIK', '']);
        exportData.push(['Token', rowData.token]);
        exportData.push(['kWh Ditambahkan', parseFloat(rowData.added_kwh || 0).toFixed(2) + ' kWh']);
    }

    // Add QR Code URL if available
    if (rowData.qr_code_url) {
        exportData.push(['', '']);
        exportData.push(['QR Code URL', rowData.qr_code_url]);
    }

    // Add completion/expiry dates if available
    if (rowData.completed_at) {
        exportData.push(['', '']);
        exportData.push(['Selesai pada', new Date(rowData.completed_at).toLocaleString('id-ID')]);
    }
    if (rowData.expired_at) {
        exportData.push(['Kadaluarsa pada', new Date(rowData.expired_at).toLocaleString('id-ID')]);
    }

    // Add additional info if available
    if (rowData.ip_address) {
        exportData.push(['', '']);
        exportData.push(['INFORMASI TAMBAHAN', '']);
        exportData.push(['IP Address', rowData.ip_address]);
    }

    // Create worksheet
    const ws = XLSX.utils.aoa_to_sheet(exportData);

    // Set column widths
    ws['!cols'] = [
        { wch: 25 },
        { wch: 40 }
    ];

    // Create workbook
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, 'Transaksi');

    // Generate filename
    const filename = `Transaksi_${rowData.transaction_id}_${new Date().toISOString().split('T')[0]}.xlsx`;

    // Download
    XLSX.writeFile(wb, filename);

    Swal.fire({
        title: 'Berhasil!',
        text: 'File Excel berhasil diunduh',
        icon: 'success',
        timer: 2000,
        showConfirmButton: false
    });
}

/**
 * Export a single transaction to PDF
 */
function exportTransactionToPDF(rowData) {
    const { jsPDF } = window.jspdf;
    const doc = new jsPDF();

    // Set font
    doc.setFont('helvetica');

    // Title
    doc.setFontSize(18);
    doc.setTextColor(40, 40, 40);
    doc.text('DETAIL TRANSAKSI', 105, 20, { align: 'center' });

    // Transaction ID
    doc.setFontSize(12);
    doc.setTextColor(100, 100, 100);
    doc.text(`ID: ${rowData.transaction_id}`, 105, 28, { align: 'center' });

    // Line separator
    doc.setLineWidth(0.5);
    doc.line(20, 35, 190, 35);

    let yPos = 45;

    // Transaction Information
    doc.setFontSize(14);
    doc.setTextColor(40, 40, 40);
    doc.text('Informasi Transaksi', 20, yPos);
    yPos += 8;

    doc.setFontSize(11);
    doc.setTextColor(60, 60, 60);
    const transactionInfo = [
        ['Tanggal Dibuat', new Date(rowData.initiated_at).toLocaleString('id-ID')],
        ['Status', rowData.status.toUpperCase()],
        ['Tipe Pembayaran', rowData.payment_type === 'prabayar' ? 'Prabayar' : 'Pascabayar'],
        ['Metode Pembayaran', getPaymentMethod(rowData.payment_method)]
    ];

    transactionInfo.forEach(([label, value]) => {
        doc.text(label + ':', 25, yPos);
        doc.text(value, 80, yPos);
        yPos += 7;
    });

    yPos += 5;

    // Customer Information
    doc.setFontSize(14);
    doc.setTextColor(40, 40, 40);
    doc.text('Informasi Pelanggan', 20, yPos);
    yPos += 8;

    doc.setFontSize(11);
    doc.setTextColor(60, 60, 60);
    const customerInfo = [
        ['ID Pelanggan', rowData.customer_id],
        ['No. Meter', rowData.meter_number || '-']
    ];

    customerInfo.forEach(([label, value]) => {
        doc.text(label + ':', 25, yPos);
        doc.text(value, 80, yPos);
        yPos += 7;
    });

    yPos += 5;

    // Cost Details
    doc.setFontSize(14);
    doc.setTextColor(40, 40, 40);
    doc.text('Rincian Biaya', 20, yPos);
    yPos += 8;

    doc.setFontSize(11);
    doc.setTextColor(60, 60, 60);
    const costDetails = [
        ['Nominal', 'Rp ' + parseInt(rowData.amount || 0).toLocaleString('id-ID')],
        ['Biaya Admin', 'Rp ' + parseInt(rowData.admin_fee || 0).toLocaleString('id-ID')],
        ['Total', 'Rp ' + parseInt(rowData.total_amount || 0).toLocaleString('id-ID')]
    ];

    costDetails.forEach(([label, value], index) => {
        doc.text(label + ':', 25, yPos);
        doc.text(value, 80, yPos);
        if (index === costDetails.length - 1) {
            doc.setFontSize(12);
            doc.setFont('helvetica', 'bold');
        }
        yPos += 7;
    });

    // Token info for prepaid
    if (rowData.payment_type === 'prabayar' && rowData.token) {
        yPos += 5;
        doc.setFont('helvetica', 'normal');
        doc.setFontSize(14);
        doc.setTextColor(40, 40, 40);
        doc.text('Token Listrik', 20, yPos);
        yPos += 8;

        doc.setFontSize(11);
        doc.setTextColor(60, 60, 60);
        doc.text('Token:', 25, yPos);
        doc.setFont('helvetica', 'bold');
        doc.text(rowData.token, 80, yPos);
        doc.setFont('helvetica', 'normal');
        yPos += 7;
        doc.text('kWh Ditambahkan:', 25, yPos);
        doc.text(parseFloat(rowData.added_kwh || 0).toFixed(2) + ' kWh', 80, yPos);
        yPos += 7;
    }

    // QR Code for payment (if available)
    if (rowData.qr_code_url) {
        yPos += 5;
        
        // Check if we need a new page
        if (yPos > 240) {
            doc.addPage();
            yPos = 20;
        }
        
        doc.setFont('helvetica', 'normal');
        doc.setFontSize(14);
        doc.setTextColor(40, 40, 40);
        doc.text('QR Code Pembayaran', 20, yPos);
        yPos += 8;

        // Add QR code image
        try {
            // Add the image to PDF (QR code is typically centered)
            doc.addImage(rowData.qr_code_url, 'PNG', 25, yPos, 50, 50);
            yPos += 55;
            
            doc.setFontSize(9);
            doc.setTextColor(100, 100, 100);
            doc.text('Scan QR Code untuk melakukan pembayaran', 25, yPos);
            yPos += 5;
        } catch (error) {
            console.error('Error adding QR code to PDF:', error);
            doc.setFontSize(10);
            doc.setTextColor(150, 150, 150);
            doc.text('QR Code tidak dapat ditampilkan', 25, yPos);
            doc.text('URL: ' + rowData.qr_code_url.substring(0, 60) + '...', 25, yPos + 5);
            yPos += 12;
        }
    }

    // Completion/Expiry dates
    yPos += 5;
    doc.setFontSize(10);
    doc.setTextColor(60, 60, 60);
    
    if (rowData.completed_at) {
        doc.text('Selesai pada: ' + new Date(rowData.completed_at).toLocaleString('id-ID'), 20, yPos);
        yPos += 6;
    }
    if (rowData.expired_at) {
        doc.text('Kadaluarsa pada: ' + new Date(rowData.expired_at).toLocaleString('id-ID'), 20, yPos);
        yPos += 6;
    }

    // Additional info
    if (rowData.ip_address) {
        yPos += 3;
        doc.setFontSize(9);
        doc.setTextColor(120, 120, 120);
        doc.text('IP Address: ' + rowData.ip_address, 20, yPos);
        yPos += 5;
    }

    // Footer
    yPos += 10;
    doc.setFontSize(9);
    doc.setTextColor(150, 150, 150);
    doc.text('Dicetak pada: ' + new Date().toLocaleString('id-ID'), 105, yPos, { align: 'center' });

    // Generate filename
    const filename = `Transaksi_${rowData.transaction_id}_${new Date().toISOString().split('T')[0]}.pdf`;

    // Download
    doc.save(filename);

    Swal.fire({
        title: 'Berhasil!',
        text: 'File PDF berhasil diunduh',
        icon: 'success',
        timer: 2000,
        showConfirmButton: false
    });
}

/**
 * Export all transactions
 * @param {string} format - 'excel' or 'pdf'
 */
function exportAllTransactions(format) {
    // Get all data from the table
    const allData = tableTransactions.rows().data().toArray();

    if (allData.length === 0) {
        Swal.fire('Perhatian!', 'Tidak ada data untuk diekspor', 'warning');
        return;
    }

    if (format === 'excel') {
        exportAllToExcel(allData);
    } else if (format === 'pdf') {
        exportAllToPDF(allData);
    }
}

/**
 * Export all transactions to Excel
 */
function exportAllToExcel(allData) {
    // Prepare headers
    const headers = [
        'ID Transaksi',
        'Tanggal',
        'ID Pelanggan',
        'No. Meter',
        'Tipe',
        'Metode Bayar',
        'Nominal',
        'Biaya Admin',
        'Total',
        'Status',
        'Token',
        'kWh Ditambahkan',
        'Selesai Pada',
        'Kadaluarsa Pada'
    ];

    // Prepare data rows
    const rows = allData.map(row => [
        row.transaction_id,
        new Date(row.initiated_at).toLocaleString('id-ID'),
        row.customer_id,
        row.meter_number,
        row.payment_type === 'prabayar' ? 'Prabayar' : 'Pascabayar',
        getPaymentMethod(row.payment_method),
        parseInt(row.amount),
        parseInt(row.admin_fee),
        parseInt(row.total_amount),
        row.status,
        row.token || '-',
        row.added_kwh ? parseFloat(row.added_kwh).toFixed(2) : '-',
        row.completed_at ? new Date(row.completed_at).toLocaleString('id-ID') : '-',
        row.expired_at ? new Date(row.expired_at).toLocaleString('id-ID') : '-'
    ]);

    // Combine headers and rows
    const exportData = [headers, ...rows];

    // Create worksheet
    const ws = XLSX.utils.aoa_to_sheet(exportData);

    // Set column widths
    ws['!cols'] = [
        { wch: 20 }, // ID Transaksi
        { wch: 20 }, // Tanggal
        { wch: 15 }, // ID Pelanggan
        { wch: 15 }, // No. Meter
        { wch: 12 }, // Tipe
        { wch: 15 }, // Metode Bayar
        { wch: 12 }, // Nominal
        { wch: 12 }, // Biaya Admin
        { wch: 12 }, // Total
        { wch: 12 }, // Status
        { wch: 20 }, // Token
        { wch: 15 }, // kWh
        { wch: 20 }, // Selesai
        { wch: 20 }  // Kadaluarsa
    ];

    // Create workbook
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, 'Riwayat Transaksi');

    // Generate filename
    const filename = `Riwayat_Transaksi_${new Date().toISOString().split('T')[0]}.xlsx`;

    // Download
    XLSX.writeFile(wb, filename);

    Swal.fire({
        title: 'Berhasil!',
        text: `${allData.length} transaksi berhasil diekspor ke Excel`,
        icon: 'success',
        timer: 2000,
        showConfirmButton: false
    });
}

/**
 * Export all transactions to PDF
 */
function exportAllToPDF(allData) {
    const { jsPDF } = window.jspdf;
    const doc = new jsPDF('l', 'mm', 'a4'); // Landscape mode

    // Title
    doc.setFontSize(18);
    doc.setTextColor(40, 40, 40);
    doc.text('RIWAYAT TRANSAKSI', 148, 15, { align: 'center' });
    doc.text('PLTMH LEMBANG PALESAN', 148, 23, { align: 'center' });

    // Date range
    doc.setFontSize(10);
    doc.setTextColor(100, 100, 100);
    doc.text('Diekspor pada: ' + new Date().toLocaleString('id-ID'), 148, 30, { align: 'center' });

    // Summary
    const totalAmount = allData.reduce((sum, row) => sum + parseInt(row.total_amount || 0), 0);
    const completedCount = allData.filter(row => row.status === 'completed').length;
    
    doc.setFontSize(9);
    doc.text(`Total Transaksi: ${allData.length} | Selesai: ${completedCount} | Total Nilai: Rp ${totalAmount.toLocaleString('id-ID')}`, 148, 36, { align: 'center' });

    // Prepare table data with better formatting
    const tableData = allData.map(row => {
        // Format status
        let statusText = row.status;
        switch(row.status) {
            case 'completed': statusText = 'Selesai'; break;
            case 'pending': statusText = 'Pending'; break;
            case 'failed': statusText = 'Gagal'; break;
            case 'expired': statusText = 'Kadaluarsa'; break;
        }

        return [
            row.transaction_id, // Show full transaction ID - let autoTable handle wrapping
            new Date(row.initiated_at).toLocaleDateString('id-ID', { 
                day: '2-digit', 
                month: '2-digit', 
                year: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
            }),
            row.customer_id,
            row.meter_number || '-',
            row.payment_type === 'prabayar' ? 'Prabayar' : 'Pascabayar',
            getPaymentMethod(row.payment_method),
            'Rp ' + parseInt(row.total_amount || 0).toLocaleString('id-ID'),
            statusText
        ];
    });

    // Create table
    doc.autoTable({
        startY: 42,
        head: [['ID Transaksi', 'Tanggal & Waktu', 'ID Pelanggan', 'No. Meter', 'Tipe', 'Metode', 'Total', 'Status']],
        body: tableData,
        theme: 'striped',
        styles: {
            fontSize: 7,
            cellPadding: 2,
            overflow: 'linebreak'
        },
        headStyles: {
            fillColor: [41, 128, 185],
            textColor: 255,
            fontStyle: 'bold',
            fontSize: 8
        },
        columnStyles: {
            0: { cellWidth: 52, fontSize: 6 }, // ID Transaksi
            1: { cellWidth: 28 }, // Tanggal
            2: { cellWidth: 30 }, // ID Pelanggan
            3: { cellWidth: 28 }, // No. Meter
            4: { cellWidth: 23 }, // Tipe
            5: { cellWidth: 25 }, // Metode
            6: { cellWidth: 30, halign: 'right' }, // Total
            7: { cellWidth: 22, halign: 'center' } // Status
        },
        alternateRowStyles: {
            fillColor: [245, 245, 245]
        },
        didDrawPage: function(data) {
            // Header on each page
            if (data.pageNumber > 1) {
                doc.setFontSize(12);
                doc.setTextColor(40, 40, 40);
                doc.text('RIWAYAT TRANSAKSI (Lanjutan)', 148, 15, { align: 'center' });
            }
        }
    });

    // Footer with page numbers and summary
    const pageCount = doc.internal.getNumberOfPages();
    for (let i = 1; i <= pageCount; i++) {
        doc.setPage(i);
        doc.setFontSize(8);
        doc.setTextColor(150);
        
        // Page number
        doc.text(
            `Halaman ${i} dari ${pageCount}`,
            doc.internal.pageSize.width / 2,
            doc.internal.pageSize.height - 10,
            { align: 'center' }
        );
        
        // Footer info
        doc.setFontSize(7);
        doc.text(
            'PLTMH Lembang Palesan - Sistem Pembayaran Listrik',
            10,
            doc.internal.pageSize.height - 10
        );
        
        doc.text(
            new Date().toLocaleDateString('id-ID'),
            doc.internal.pageSize.width - 10,
            doc.internal.pageSize.height - 10,
            { align: 'right' }
        );
    }

    // Generate filename
    const filename = `Riwayat_Transaksi_${new Date().toISOString().split('T')[0]}.pdf`;

    // Download
    doc.save(filename);

    Swal.fire({
        title: 'Berhasil!',
        text: `${allData.length} transaksi berhasil diekspor ke PDF`,
        icon: 'success',
        timer: 2000,
        showConfirmButton: false
    });
}
