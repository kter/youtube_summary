import { useState, useEffect } from 'react';
import { Youtube, RefreshCw, Sparkles, User } from 'lucide-react';
import VideoCard from './components/VideoCard';
import LoadingSkeleton from './components/LoadingSkeleton';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

function App() {
  const [summaries, setSummaries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [channelName] = useState('@noiehoie');

  // Fetch summaries on mount
  useEffect(() => {
    fetchSummaries();
  }, []);

  const fetchSummaries = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/summaries`);
      if (!response.ok) throw new Error('Failed to fetch summaries');
      const data = await response.json();
      setSummaries(data.summaries);
    } catch (err) {
      setError(err.message);
      setSummaries([]);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    fetchSummaries();
  };

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="glass sticky top-0 z-50 shadow-lg">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-xl bg-gradient-to-br from-primary-500 to-accent-500">
                <Youtube className="h-6 w-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold gradient-text">YouTube Summary</h1>
                <p className="text-xs text-slate-400">動画要約サービス</p>
              </div>
            </div>
            <button
              onClick={handleRefresh}
              disabled={loading}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-primary-600 hover:bg-primary-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              <span className="hidden sm:inline">更新</span>
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Channel Info */}
        <div className="glass-card rounded-xl p-4 mb-8 flex items-center gap-3">
          <div className="p-2 rounded-full bg-gradient-to-br from-red-500 to-pink-500">
            <User className="h-5 w-5 text-white" />
          </div>
          <div>
            <h2 className="text-lg font-semibold text-slate-200">{channelName}</h2>
            <p className="text-sm text-slate-400">このチャンネルの動画を自動要約しています</p>
          </div>
        </div>

        {/* Content Section */}
        <section>
          {/* Section Header */}
          <div className="flex items-center gap-2 mb-6">
            <Sparkles className="h-5 w-5 text-accent-400" />
            <h2 className="text-lg font-semibold text-slate-200">
              動画要約一覧
            </h2>
            <span className="text-sm text-slate-500">
              ({summaries.length}件)
            </span>
          </div>

          {/* Error State */}
          {error && (
            <div className="glass-card rounded-xl p-6 text-center">
              <p className="text-red-400">エラーが発生しました: {error}</p>
              <button
                onClick={handleRefresh}
                className="mt-4 px-4 py-2 bg-primary-600 hover:bg-primary-500 rounded-lg transition-colors"
              >
                再試行
              </button>
            </div>
          )}

          {/* Loading State */}
          {loading && !error && (
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {[...Array(6)].map((_, i) => (
                <LoadingSkeleton key={i} />
              ))}
            </div>
          )}

          {/* Empty State */}
          {!loading && !error && summaries.length === 0 && (
            <div className="glass-card rounded-xl p-12 text-center">
              <Youtube className="h-16 w-16 text-slate-600 mx-auto mb-4" />
              <h3 className="text-xl font-semibold text-slate-300 mb-2">
                まだ要約がありません
              </h3>
              <p className="text-slate-500">
                このチャンネルの動画要約は、次回のバッチ処理で生成されます。
              </p>
            </div>
          )}

          {/* Video Cards Grid */}
          {!loading && !error && summaries.length > 0 && (
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {summaries.map((summary, index) => (
                <VideoCard
                  key={summary.videoId}
                  summary={summary}
                  style={{ animationDelay: `${index * 50}ms` }}
                />
              ))}
            </div>
          )}
        </section>
      </main>

      {/* Footer */}
      <footer className="mt-16 py-8 border-t border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center text-slate-500 text-sm">
          <p>Powered by Gemini 2.5 Flash</p>
          <p className="mt-1">© 2024 YouTube Summary Service</p>
        </div>
      </footer>
    </div>
  );
}

export default App;
