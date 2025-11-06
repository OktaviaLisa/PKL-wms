import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import '../services/api_service.dart';
import 'returns_screen.dart';

class QualityControlScreen extends StatefulWidget {
  const QualityControlScreen({super.key});

  @override
  State<QualityControlScreen> createState() => _QualityControlScreenState();
}

class _QualityControlScreenState extends State<QualityControlScreen> {
  List<Map<String, dynamic>> _qualityChecks = [];
  bool _isLoading = true;
  final ApiService _apiService = ApiService();

  @override
  void initState() {
    super.initState();
    _loadQualityChecks();
  }

  Future<void> _loadQualityChecks() async {
    try {
      final existingQCs = await _apiService.getQualityChecks();
      final qcMap = <int, Map<String, dynamic>>{};
      for (var qc in existingQCs) {
        qcMap[qc['reception_id']] = qc;
      }

      final receptions = await _apiService.getReceptions();
      final qualityCheckItems = <Map<String, dynamic>>[];

      for (var item in receptions) {
        if (item['status'] == 'quality_check') {
          final existingQC = qcMap[item['id']];
          qualityCheckItems.add({
            'id': item['id'],
            'product_name': item['product_name'],
            'supplier': item['supplier'],
            'quantity': item['quantity'],
            'location': item['location'],
            'notes': item['notes'],
            'date': item['date'],
            'status': existingQC != null ? existingQC['status'] : 'PENDING',
            'check_type': 'INCOMING',
          });
        }
      }

      setState(() {
        _qualityChecks = qualityCheckItems;
        _isLoading = false;
      });
      print('✅ QC data loaded: ${_qualityChecks.length} items');
    } catch (e) {
      print('❌ Error loading QC data: $e');
      setState(() {
        _qualityChecks = [];
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Quality Control'),
        backgroundColor: Colors.red[700],
        foregroundColor: Colors.white,
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _qualityChecks.isEmpty
              ? _buildEmptyState()
              : ListView.builder(
                  padding: const EdgeInsets.all(8),
                  itemCount: _qualityChecks.length,
                  itemBuilder: (context, index) {
                    final check = _qualityChecks[index];
                    return Card(
                      margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 8),
                      child: ListTile(
                        leading: CircleAvatar(
                          backgroundColor: Colors.red[100],
                          child: Icon(Icons.verified, color: Colors.red[700]),
                        ),
                        title: Text(check['product_name'] ?? 'Produk Tidak Diketahui'),
                        subtitle: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Supplier: ${check['supplier'] ?? 'N/A'}'),
                            Text('Jumlah: ${check['quantity']} unit'),
                            Text('Lokasi: ${check['location'] ?? 'N/A'}'),
                            Text('Tanggal: ${check['date'] ?? 'N/A'}'),
                          ],
                        ),
                        trailing: _buildStatusChip(check['status'] ?? 'PENDING'),
                        isThreeLine: true,
                        onTap: () => _showQualityCheckDialog(check),
                      ),
                    );
                  },
                ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.verified_outlined, size: 64, color: Colors.grey[400]),
          const SizedBox(height: 16),
          Text('Belum ada barang untuk Quality Control',
              style: TextStyle(fontSize: 18, color: Colors.grey[600])),
          const SizedBox(height: 8),
          Text('Tambahkan barang di Penerimaan Barang terlebih dahulu',
              style: TextStyle(color: Colors.grey[500])),
          const SizedBox(height: 16),
          ElevatedButton(
            onPressed: () => _loadQualityChecks(),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red[700],
            ),
            child: const Text('Refresh', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  Widget _buildStatusChip(String status) {
    Color color;
    String displayText;
    switch (status) {
      case 'PASS':
        color = Colors.green;
        displayText = 'Lulus';
        break;
      case 'FAIL':
        color = Colors.red;
        displayText = 'Gagal';
        break;
      case 'PENDING':
      default:
        color = Colors.orange;
        displayText = 'Menunggu';
    }
    return Chip(
      label: Text(displayText, style: TextStyle(color: color, fontSize: 12)),
      backgroundColor: color.withAlpha(25),
    );
  }

  void _showQualityCheckDialog(Map<String, dynamic> item) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Quality Check'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Produk: ${item['product_name']}',
                style: const TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            Text('Supplier: ${item['supplier'] ?? 'N/A'}'),
            Text('Jumlah: ${item['quantity']} unit'),
            Text('Lokasi: ${item['location'] ?? 'N/A'}'),
            if (item['notes'] != null && item['notes'].toString().isNotEmpty)
              Text('Catatan: ${item['notes']}'),
            const SizedBox(height: 20),
            const Text('Pilih hasil Quality Check:',
                style: TextStyle(fontWeight: FontWeight.w600)),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: () {
                      _updateQCStatus(item, 'PASS');
                      Navigator.pop(context);
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.green,
                      padding: const EdgeInsets.symmetric(vertical: 12),
                    ),
                    icon: const Icon(Icons.check_circle, color: Colors.white),
                    label: const Text('LULUS', style: TextStyle(color: Colors.white)),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: () {
                      Navigator.pop(context);
                      _handleFailedQC(item);
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.red,
                      padding: const EdgeInsets.symmetric(vertical: 12),
                    ),
                    icon: const Icon(Icons.cancel, color: Colors.white),
                    label: const Text('GAGAL', style: TextStyle(color: Colors.white)),
                  ),
                ),
              ],
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
        ],
      ),
    );
  }

  void _updateQCStatus(Map<String, dynamic> item, String status) async {
    try {
      await _apiService.createQualityCheckRecord({
        'reception_id': item['id'],
        'product_name': item['product_name'],
        'quantity': item['quantity'],
        'status': status,
        'notes': 'QC performed via mobile app',
      });

      await _loadQualityChecks();

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
              '${item['product_name']} - Lulus QC dan disimpan ke inventory'),
          backgroundColor: Colors.green,
        ),
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Gagal menyimpan hasil QC: $e'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  void _handleFailedQC(Map<String, dynamic> item) async {
    try {
      print('🔄 Starting failed QC process for: ${item['product_name']}');
      print('📋 Item data: $item');
      
      // Update QC status to FAIL
      print('📝 Creating QC record...');
      final qcResult = await _apiService.createQualityCheckRecord({
        'reception_id': item['id'],
        'product_name': item['product_name'],
        'quantity': item['quantity'],
        'status': 'FAIL',
        'notes': 'Failed QC - redirected to returns',
      });
      print('✅ QC record created: $qcResult');

      // Create return record for failed QC item
      print('📝 Creating return record...');
      final returnData = {
        'product_name': item['product_name'],
        'quantity': item['quantity'],
        'return_type': 'QUALITY_FAIL',
        'reason': 'Failed quality control inspection',
        'supplier': item['supplier'] ?? 'N/A',
        'reception_id': item['id'],
      };
      print('📋 Return data: $returnData');
      
      final returnResult = await _apiService.createReturn(returnData);
      print('✅ Return record created: $returnResult');

      await _loadQualityChecks();

      // Navigate to returns screen
      Navigator.push(
        context,
        MaterialPageRoute(builder: (context) => const ReturnsScreen()),
      );
      
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('${item['product_name']} - Gagal QC, disimpan ke Retur'),
          backgroundColor: Colors.orange,
        ),
      );
    } catch (e) {
      print('❌ Error in _handleFailedQC: $e');
      print('📋 Stack trace: ${StackTrace.current}');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Gagal memproses QC: $e'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  void _showAddQualityCheckDialog() {}
}
