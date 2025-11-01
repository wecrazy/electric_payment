// PLTMH Customer Management JavaScript
let tablePrepaid, tablePostpaid;

$(document).ready(function() {
    initializeTables();
});

function initializeTables() {
    // Initialize Prepaid Customers Table
    tablePrepaid = $('#tablePrepaid').DataTable({
        processing: true,
        serverSide: false,
        ajax: {
            url: PLTMH_ENDPOINTS.prepaid.table,
            type: 'POST',
            headers: {
                'X-CSRF-TOKEN': $('meta[name="csrf-token"]').attr('content')
            },
            dataSrc: 'data'
        },
        columns: [
            { data: 'customer_id' },
            { data: 'meter_number' },
            { data: 'tariff_code' },
            { data: 'power_va' },
            { 
                data: 'balance_kwh',
                render: function(data) {
                    return parseFloat(data).toFixed(2) + ' kWh';
                }
            },
            { 
                data: 'last_token',
                render: function(data) {
                    return data || '-';
                }
            },
            { 
                data: 'last_topup_at',
                render: function(data) {
                    return data ? new Date(data).toLocaleString('id-ID') : '-';
                }
            },
            {
                data: null,
                orderable: false,
                render: function(data, type, row) {
                    return `
                        <button class="btn btn-sm btn-info" onclick="viewPrepaidDetail('${row.customer_id}')">
                            <i class="bx bx-show"></i>
                        </button>
                        <button class="btn btn-sm btn-warning" onclick="editPrepaid(${row.ID})">
                            <i class="bx bx-edit"></i>
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="deletePrepaid(${row.ID}, '${row.customer_id}')">
                            <i class="bx bx-trash"></i>
                        </button>
                    `;
                }
            }
        ],
        order: [[0, 'asc']],
        language: {
            url: '/assets/self/datatables/id.json'
        }
    });

    // Initialize Postpaid Customers Table
    tablePostpaid = $('#tablePostpaid').DataTable({
        processing: true,
        serverSide: false,
        ajax: {
            url: PLTMH_ENDPOINTS.postpaid.table,
            type: 'POST',
            headers: {
                'X-CSRF-TOKEN': $('meta[name="csrf-token"]').attr('content')
            },
            dataSrc: 'data'
        },
        columns: [
            { data: 'customer_id' },
            { data: 'meter_number' },
            { data: 'tariff_code' },
            { data: 'power_va' },
            { 
                data: 'current_usage_kwh',
                render: function(data) {
                    return parseFloat(data).toFixed(2) + ' kWh';
                }
            },
            { 
                data: 'outstanding_balance',
                render: function(data) {
                    return 'Rp ' + parseInt(data).toLocaleString('id-ID');
                }
            },
            {
                data: 'is_disconnected',
                render: function(data) {
                    return data ? '<span class="badge bg-danger">Terputus</span>' : '<span class="badge bg-success">Aktif</span>';
                }
            },
            {
                data: null,
                orderable: false,
                render: function(data, type, row) {
                    return `
                        <button class="btn btn-sm btn-info" onclick="viewPostpaidDetail('${row.customer_id}')">
                            <i class="bx bx-show"></i>
                        </button>
                        <button class="btn btn-sm btn-warning" onclick="editPostpaid(${row.ID})">
                            <i class="bx bx-edit"></i>
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="deletePostpaid(${row.ID}, '${row.customer_id}')">
                            <i class="bx bx-trash"></i>
                        </button>
                    `;
                }
            }
        ],
        order: [[0, 'asc']],
        language: {
            url: '/assets/self/datatables/id.json'
        }
    });
}

// ===========================
// PREPAID CUSTOMER FUNCTIONS
// ===========================

function submitAddPrepaid() {
    const form = document.getElementById('formAddPrepaid');
    const formData = new FormData(form);
    const data = Object.fromEntries(formData);
    data.power_va = parseInt(data.power_va);

    $.ajax({
        url: PLTMH_ENDPOINTS.prepaid.create,
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            if (response.success) {
                Swal.fire('Berhasil!', response.message, 'success');
                $('#modalAddPrepaid').modal('hide');
                form.reset();
                tablePrepaid.ajax.reload();
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

function editPrepaid(id) {
    // Get customer data from table
    const rowData = tablePrepaid.rows().data().toArray().find(row => row.ID === id);
    
    if (!rowData) {
        Swal.fire('Kesalahan!', 'Data tidak ditemukan', 'error');
        return;
    }

    // Fill form
    const form = document.getElementById('formEditPrepaid');
    form.querySelector('[name="id"]').value = rowData.ID;
    form.querySelector('[name="customer_id"]').value = rowData.customer_id;
    form.querySelector('[name="meter_number"]').value = rowData.meter_number;
    form.querySelector('[name="tariff_code"]').value = rowData.tariff_code;
    form.querySelector('[name="power_va"]').value = rowData.power_va;
    form.querySelector('[name="balance_kwh"]').value = rowData.balance_kwh;

    $('#modalEditPrepaid').modal('show');
}

function submitEditPrepaid() {
    const form = document.getElementById('formEditPrepaid');
    const formData = new FormData(form);
    const data = Object.fromEntries(formData);
    data.id = parseInt(data.id);
    data.power_va = parseInt(data.power_va);
    data.balance_kwh = parseFloat(data.balance_kwh);

    $.ajax({
        url: PLTMH_ENDPOINTS.prepaid.update,
        type: 'PUT',
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            if (response.success) {
                Swal.fire('Berhasil!', response.message, 'success');
                $('#modalEditPrepaid').modal('hide');
                tablePrepaid.ajax.reload();
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

function deletePrepaid(id, customerId) {
    Swal.fire({
        title: 'Hapus Pelanggan?',
        text: `Anda yakin ingin menghapus pelanggan ${customerId}?`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#d33',
        cancelButtonColor: '#3085d6',
        confirmButtonText: 'Ya, Hapus!',
        cancelButtonText: 'Batal'
    }).then((result) => {
        if (result.isConfirmed) {
            $.ajax({
                url: PLTMH_ENDPOINTS.prepaid.delete + id,
                type: 'DELETE',
                success: function(response) {
                    if (response.success) {
                        Swal.fire('Terhapus!', response.message, 'success');
                        tablePrepaid.ajax.reload();
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
    });
}

function viewPrepaidDetail(customerId) {
    $.ajax({
        url: PLTMH_ENDPOINTS.customer.detail + customerId + '/detail?type=prepaid',
        type: 'GET',
        success: function(response) {
            if (response.success) {
                const data = response.data;
                const html = `
                    <div class="mb-3">
                        <h6>Informasi Pelanggan</h6>
                        <table class="table table-borderless">
                            <tr><td><strong>ID Pelanggan:</strong></td><td>${data.customer_id}</td></tr>
                            <tr><td><strong>No. Meter:</strong></td><td>${data.meter_number}</td></tr>
                            <tr><td><strong>Kode Tarif:</strong></td><td>${data.tariff_code}</td></tr>
                            <tr><td><strong>Daya:</strong></td><td>${data.power_va} VA</td></tr>
                            <tr><td><strong>Saldo kWh:</strong></td><td>${parseFloat(data.balance_kwh).toFixed(2)} kWh</td></tr>
                            <tr><td><strong>Token Terakhir:</strong></td><td>${data.last_token || '-'}</td></tr>
                        </table>
                    </div>
                    ${data.TopUpHistory && data.TopUpHistory.length > 0 ? `
                        <div class="mb-3">
                            <h6>Riwayat Top-Up (10 Terakhir)</h6>
                            <div class="table-responsive">
                                <table class="table table-sm">
                                    <thead>
                                        <tr>
                                            <th>Tanggal</th>
                                            <th>Token</th>
                                            <th>Nominal</th>
                                            <th>kWh</th>
                                            <th>Vendor</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        ${data.TopUpHistory.map(h => `
                                            <tr>
                                                <td>${new Date(h.top_up_at).toLocaleString('id-ID')}</td>
                                                <td><code>${h.token}</code></td>
                                                <td>Rp ${parseInt(h.amount_rp).toLocaleString('id-ID')}</td>
                                                <td>${parseFloat(h.added_kwh).toFixed(2)} kWh</td>
                                                <td>${h.vendor || '-'}</td>
                                            </tr>
                                        `).join('')}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    ` : ''}
                `;
                
                Swal.fire({
                    title: 'Detail Pelanggan Prabayar',
                    html: html,
                    width: 800,
                    confirmButtonText: 'Tutup'
                });
            }
        },
        error: function(xhr) {
              Swal.fire('Kesalahan!', 'Gagal memuat detail pelanggan', 'error');
        }
    });
}

// ===========================
// POSTPAID CUSTOMER FUNCTIONS
// ===========================

function submitAddPostpaid() {
    const form = document.getElementById('formAddPostpaid');
    const formData = new FormData(form);
    const data = Object.fromEntries(formData);
    data.power_va = parseInt(data.power_va);

    $.ajax({
        url: PLTMH_ENDPOINTS.postpaid.create,
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            if (response.success) {
                Swal.fire('Berhasil!', response.message, 'success');
                $('#modalAddPostpaid').modal('hide');
                form.reset();
                tablePostpaid.ajax.reload();
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

function editPostpaid(id) {
    // Get customer data from table
    const rowData = tablePostpaid.rows().data().toArray().find(row => row.ID === id);
    
    if (!rowData) {
           Swal.fire('Kesalahan!', 'Data tidak ditemukan', 'error');
        return;
    }

    // Fill form
    const form = document.getElementById('formEditPostpaid');
    form.querySelector('[name="id"]').value = rowData.ID;
    form.querySelector('[name="customer_id"]').value = rowData.customer_id;
    form.querySelector('[name="meter_number"]').value = rowData.meter_number;
    form.querySelector('[name="tariff_code"]').value = rowData.tariff_code;
    form.querySelector('[name="power_va"]').value = rowData.power_va;
    form.querySelector('[name="current_usage_kwh"]').value = rowData.current_usage_kwh;
    form.querySelector('[name="outstanding_balance"]').value = rowData.outstanding_balance;
    form.querySelector('[name="is_disconnected"]').checked = rowData.is_disconnected;

    $('#modalEditPostpaid').modal('show');
}

function submitEditPostpaid() {
    const form = document.getElementById('formEditPostpaid');
    const formData = new FormData(form);
    const data = Object.fromEntries(formData);
    data.id = parseInt(data.id);
    data.power_va = parseInt(data.power_va);
    data.current_usage_kwh = parseFloat(data.current_usage_kwh);
    data.outstanding_balance = parseInt(data.outstanding_balance);
    data.is_disconnected = form.querySelector('[name="is_disconnected"]').checked;

    $.ajax({
        url: PLTMH_ENDPOINTS.postpaid.update,
        type: 'PUT',
        contentType: 'application/json',
        data: JSON.stringify(data),
        success: function(response) {
            if (response.success) {
                Swal.fire('Berhasil!', response.message, 'success');
                $('#modalEditPostpaid').modal('hide');
                tablePostpaid.ajax.reload();
            } else {
                Swal.fire('Gagal!', response.message, 'error');
            }
        },
        error: function(xhr) {
            const error = xhr.responseJSON ? xhr.responseJSON.message : 'Terjadi kesalahan';
            Swal.fire('Error!', error, 'error');
        }
    });
}

function deletePostpaid(id, customerId) {
    Swal.fire({
        title: 'Hapus Pelanggan?',
        text: `Anda yakin ingin menghapus pelanggan ${customerId}?`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#d33',
        cancelButtonColor: '#3085d6',
        confirmButtonText: 'Ya, Hapus!',
        cancelButtonText: 'Batal'
    }).then((result) => {
        if (result.isConfirmed) {
            $.ajax({
                url: PLTMH_ENDPOINTS.postpaid.delete + id,
                type: 'DELETE',
                success: function(response) {
                    if (response.success) {
                        Swal.fire('Terhapus!', response.message, 'success');
                        tablePostpaid.ajax.reload();
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
    });
}

function viewPostpaidDetail(customerId) {
    $.ajax({
        url: PLTMH_ENDPOINTS.customer.detail + customerId + '/detail?type=postpaid',
        type: 'GET',
        success: function(response) {
            if (response.success) {
                const data = response.data;
                const html = `
                    <table class="table table-borderless">
                        <tr><td><strong>ID Pelanggan:</strong></td><td>${data.customer_id}</td></tr>
                        <tr><td><strong>No. Meter:</strong></td><td>${data.meter_number}</td></tr>
                        <tr><td><strong>Kode Tarif:</strong></td><td>${data.tariff_code}</td></tr>
                        <tr><td><strong>Daya:</strong></td><td>${data.power_va} VA</td></tr>
                        <tr><td><strong>Periode Tagihan:</strong></td><td>${new Date(data.billing_cycle_start).toLocaleDateString('id-ID')} - ${new Date(data.billing_cycle_end).toLocaleDateString('id-ID')}</td></tr>
                        <tr><td><strong>Pemakaian Bulan Ini:</strong></td><td>${parseFloat(data.current_usage_kwh).toFixed(2)} kWh</td></tr>
                        <tr><td><strong>Tagihan:</strong></td><td>Rp ${parseInt(data.outstanding_balance).toLocaleString('id-ID')}</td></tr>
                        <tr><td><strong>Pembayaran Terakhir:</strong></td><td>${data.last_payment_at ? new Date(data.last_payment_at).toLocaleString('id-ID') : '-'}</td></tr>
                        <tr><td><strong>Status:</strong></td><td>${data.is_disconnected ? '<span class="badge bg-danger">Terputus</span>' : '<span class="badge bg-success">Aktif</span>'}</td></tr>
                    </table>
                `;
                
                Swal.fire({
                    title: 'Detail Pelanggan Pascabayar',
                    html: html,
                    width: 800,
                    confirmButtonText: 'Tutup'
                });
            }
        },
        error: function(xhr) {
              Swal.fire('Kesalahan!', 'Gagal memuat detail pelanggan', 'error');
        }
    });
}
