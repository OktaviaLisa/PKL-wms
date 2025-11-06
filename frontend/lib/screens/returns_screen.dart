import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../services/api_service.dart';

class ReturnsScreen extends StatefulWidget {
  const ReturnsScreen({super.key});

  @override
  State<ReturnsScreen> createState() => _ReturnsScreenState();
}

class _ReturnsScreenState extends State<ReturnsScreen> {
  List<Map<String, dynamic>> _returns = [];
  bool _isLoading = true;
  final ApiService _apiService = ApiService();

  // Controller dan variabel untuk input dialog
  final TextEditingController _quantityController = TextEditingController();
  final TextEditingController _reasonController = TextEditingController();
  String? _selectedProduct;
  String? _returnType;

  @override
  void initState() {
    super.initState();
    _loadReturns();
  }

  @override
  void dispose() {
    _quantityController.dispose();
    _reasonController.dispose();
    super.dispose();
  }

  // Ambil data returns dari API
  Future<void> _loadReturns() async {
    setState(() {
      _isLoading = true;
    });
    
    try {
      print('🔄 Loading returns data...');
      final returns = await _apiService.getReturns();
      print('✅ Returns loaded: ${returns.length} items');
      
      setState(() {
        _returns = List<Map<String, dynamic>>.from(returns);
        _isLoading = false;
      });
    } catch (e) {
      print('❌ Error loading returns: $e');
      setState(() {
        _returns = [];
        _isLoading = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Gagal memuat data: $e')),
      );
    }
  }

  // Tambah return baru ke server
  Future<void> _addReturn() async {
    if (_selectedProduct == null ||
        _returnType == null ||
        _quantityController.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Lengkapi semua data!')),
      );
      return;
    }

    final newReturn = {
      "product_name": _selectedProduct,
      "quantity": int.parse(_quantityController.text),
      "return_type": _returnType,
      "reason": _reasonController.text,
      "status": "PENDING",
    };

    try {
      await _apiService.createReturn(newReturn);
      Navigator.pop(context);
      _loadReturns();
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Data pengembalian berhasil ditambahkan')),
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Gagal menyimpan data: $e')),
      );
    }
  }

  // Tampilan utama
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Pengembalian Barang'),
        backgroundColor: Colors.amber[700],
        foregroundColor: Colors.white,
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _returns.isEmpty
              ? _buildEmptyState()
              : RefreshIndicator(
                  onRefresh: _loadReturns,
                  child: ListView.builder(
                    padding: const EdgeInsets.all(8),
                    itemCount: _returns.length,
                    itemBuilder: (context, index) {
                    final returnItem = _returns[index];
                    return Card(
                      margin:
                          const EdgeInsets.symmetric(vertical: 4, horizontal: 8),
                      child: ListTile(
                        leading: CircleAvatar(
                          backgroundColor: Colors.amber[100],
                          child: Icon(Icons.keyboard_return,
                              color: Colors.amber[700]),
                        ),
                        title: Text(returnItem['product_name'] ?? '-'),
                        subtitle: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Jumlah: ${returnItem['quantity'] ?? '-'} unit'),
                            Text('Tipe: ${returnItem['return_type'] ?? '-'}'),
                            if (returnItem['supplier'] != null && returnItem['supplier'].toString().isNotEmpty)
                              Text('Supplier: ${returnItem['supplier']}'),
                            Text('Alasan: ${returnItem['reason'] ?? '-'}'),
                            Text(
                              'Tanggal: ${returnItem['created_at']?.toString().substring(0, 10) ?? 'N/A'}',
                            ),
                          ],
                        ),
                        trailing: _buildStatusChip(returnItem['status'] ?? 'PENDING'),
                        isThreeLine: true,
                      ),
                    );
                  },
                ),
              ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _showAddReturnDialog,
        backgroundColor: Colors.amber[700],
        foregroundColor: Colors.white,
        icon: const Icon(Icons.add),
        label: const Text('Tambah Return'),
      ),
    );
  }

  // Chip status
  Widget _buildStatusChip(String status) {
    Color color;
    String displayText;
    
    switch (status.toUpperCase()) {
      case 'APPROVED':
        color = Colors.green;
        displayText = 'Disetujui';
        break;
      case 'REJECTED':
        color = Colors.red;
        displayText = 'Ditolak';
        break;
      case 'PENDING':
      default:
        color = Colors.orange;
        displayText = 'Menunggu';
    }
    
    return Chip(
      label: Text(
        displayText,
        style: TextStyle(color: color, fontSize: 12),
      ),
      backgroundColor: color.withAlpha(25),
    );
  }

  // Jika tidak ada data
  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.keyboard_return, size: 64, color: Colors.grey[400]),
          const SizedBox(height: 16),
          Text(
            'Belum ada data pengembalian',
            style: TextStyle(fontSize: 18, color: Colors.grey[600]),
          ),
          const SizedBox(height: 16),
          ElevatedButton(
            onPressed: _loadReturns,
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.amber[700],
            ),
            child:
                const Text('Refresh', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  // Dialog tambah data
  void _showAddReturnDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Tambah Pengembalian'),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              DropdownButtonFormField<String>(
                decoration: const InputDecoration(labelText: 'Produk'),
                items: ['Laptop Dell XPS 13', 'Mouse Wireless', 'Keyboard']
                    .map((e) =>
                        DropdownMenuItem(value: e, child: Text(e)))
                    .toList(),
                onChanged: (value) => _selectedProduct = value,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _quantityController,
                decoration: const InputDecoration(labelText: 'Jumlah'),
                keyboardType: TextInputType.number,
                inputFormatters: [FilteringTextInputFormatter.digitsOnly],
              ),
              const SizedBox(height: 16),
              DropdownButtonFormField<String>(
                decoration: const InputDecoration(labelText: 'Tipe Return'),
                items: ['CUSTOMER', 'SUPPLIER']
                    .map((e) =>
                        DropdownMenuItem(value: e, child: Text(e)))
                    .toList(),
                onChanged: (value) => _returnType = value,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _reasonController,
                decoration: const InputDecoration(labelText: 'Alasan'),
                maxLines: 2,
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: _addReturn,
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.amber[700],
            ),
            child:
                const Text('Simpan', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }
}
