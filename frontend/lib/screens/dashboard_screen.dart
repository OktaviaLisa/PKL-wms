import 'package:flutter/material.dart';
import 'notification_screen.dart';
import 'profile_screen.dart';
import '../services/auth_service.dart';
import 'login_screen.dart';
import 'transaction_list_screen.dart';
import 'reception_form_screen.dart';
import 'dispatch_form_screen.dart';
import 'reception_screen.dart';
import 'dispatch_screen.dart';
import 'inventory_screen.dart';
import 'quality_control_screen.dart';

class DashboardScreen extends StatefulWidget {
  final String? username;

  const DashboardScreen({super.key, this.username});

  @override
  DashboardScreenState createState() => DashboardScreenState();
}

class DashboardScreenState extends State<DashboardScreen> {
  int _currentIndex = 1;
  final AuthService _authService = AuthService();
  Map<String, dynamic>? userInfo;
  List<String> userRoles = [];

  @override
  void initState() {
    super.initState();
    _loadUserInfo();
  }

  Future<void> _loadUserInfo() async {
    final info = await _authService.getUserInfo();
    if (mounted) {
      setState(() {
        userInfo = info;
        userRoles = info?['roles'] ?? [info?['role'] ?? 'user'];
      });
    }
  }

  Future<void> _logout() async {
    await _authService.logout();
    if (mounted) {
      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (context) => const LoginScreen()),
      );
    }
  }

  void _handleMenuTap(String title) {
    switch (title) {
      case 'Penerimaan':
        Navigator.push(
          context,
          MaterialPageRoute(builder: (context) => const ReceptionScreen()),
        );
        break;
      case 'Pengeluaran':
        Navigator.push(
          context,
          MaterialPageRoute(builder: (context) => const DispatchScreen()),
        );
        break;
      case 'Inventory':
        Navigator.push(
          context,
          MaterialPageRoute(builder: (context) => const InventoryScreen()),
        );
        break;
      case 'Transaksi ':
        Navigator.push(
          context,
          MaterialPageRoute(builder: (context) => const TransactionListScreen()),
        );
        break;
      case 'Validasi Barang':
        Navigator.push(
          context,
          MaterialPageRoute(builder: (context) => const QualityControlScreen()),
        );
        break;
      case 'Retur':
        Navigator.pushNamed(context, '/returns');
        break;
      default:
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('$title belum tersedia')),
        );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _currentIndex == 0
          ? const NotificationScreen()
          : _currentIndex == 1
              ? _buildDashboard()
              : ProfileScreen(username: widget.username ?? 'User'),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _currentIndex,
        onTap: (index) => setState(() => _currentIndex = index),
        backgroundColor: Colors.red[700],
        selectedItemColor: Colors.white,
        unselectedItemColor: Colors.white70,
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.notifications),
            label: 'Notifikasi',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.home),
            label: 'Beranda',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.person),
            label: 'Profil',
          ),
        ],
      ),
    );
  }

  Widget _buildDashboard() {
    return SafeArea(
      child: SingleChildScrollView(
        child: Column(
          children: [
            // Warehouse header with background image
            Container(
              height: 200,
              width: double.infinity,
              decoration: const BoxDecoration(
                image: DecorationImage(
                  image: AssetImage('assets/dashboard.jpg'),
                  fit: BoxFit.cover,
                ),
              ),
              child: Container(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    colors: [
                      Colors.black.withValues(alpha: 0.1),
                      Colors.black.withValues(alpha: 0.2),
                    ],
                  ),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 70, vertical: 14),
                        decoration: BoxDecoration(
                          color: Colors.red[700],
                          borderRadius: BorderRadius.circular(16),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(Icons.warehouse,
                                color: Colors.white, size: 20),
                            const SizedBox(width: 8),
                            const Text(
                              'Warehouse  ',
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const SizedBox(width: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 12, vertical: 4),
                              decoration: BoxDecoration(
                                color: Colors.green,
                                borderRadius: BorderRadius.circular(10),
                              ),
                              child: const Text(
                                'Active',
                                style: TextStyle(
                                  color: Colors.white,
                                  fontSize: 12,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),

            
            // Main menu horizontal ListView
            SizedBox(
            height: 150,
            width: double.infinity,
            child: Center( // ini yang bikin seluruh listView ke tengah
              child: ListView(
                shrinkWrap: true, // biar ukurannya pas konten
                scrollDirection: Axis.horizontal,
                physics: const BouncingScrollPhysics(),
                children: [
                  _buildMenuCard('Inventory', Icons.inventory_2_outlined, Colors.red),
                  _buildMenuCard('Transaksi ', Icons.receipt_long, Colors.red),
                  _buildMenuCard('Validasi Barang', Icons.security_outlined, Colors.red),
                  _buildMenuCard('Pindah Rak', Icons.mail_outline, Colors.red),
                  _buildMenuCard('Penerimaan', Icons.arrow_downward, Colors.red),
                  _buildMenuCard('Pengeluaran', Icons.arrow_upward, Colors.red),
                  _buildMenuCard('Retur', Icons.refresh, Colors.red),
                  _buildMenuCard('Lokasi Rak', Icons.grid_view, Colors.red),
                ]
                    .map(
                      (card) => Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 8),
                        child: card,
                      ),
                    )
                    .toList(),
              ),
            ),
          ),



            // Charts section
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: _buildChartCard('Penjualan', Colors.blue),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: _buildChartCard('Retur', Colors.cyan),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  _buildChartCard('Keuntungan', Colors.green, isWide: true),
                ],
              ),
            ),

            const SizedBox(height: 100),
          ],
        ),
      ),
    );
  }

  Widget _buildMenuCard(String title, IconData icon, Color color) {
    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: () => _handleMenuTap(title),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          // Ikon dalam card bundar
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              shape: BoxShape.rectangle,
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                  color: Colors.grey.withValues(alpha: 0.2),
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Icon(
              icon,
              color: color,
              size: 32,
            ),
          ),
          const SizedBox(height: 8),
          // Label di luar card
          Text(
            title,
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w500,
              color: Colors.grey[800],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildChartCard(String title, Color color, {bool isWide = false}) {
    return Container(
      height: isWide ? 120 : 100,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.red[700],
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 14,
              fontWeight: FontWeight.bold,
            ),
          ),
          Expanded(
            child: Center(
              child: Icon(
                Icons.bar_chart,
                color: Colors.white.withValues(alpha: 0.7),
                size: 40,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
