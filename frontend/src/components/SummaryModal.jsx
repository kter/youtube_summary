import { X } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import remarkBreaks from 'remark-breaks';

function SummaryModal({ summary, isOpen, onClose }) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black/70 backdrop-blur-sm animate-fade-in"
        onClick={onClose}
      />

      {/* Modal Content */}
      <div className="relative w-full max-w-3xl max-h-[90vh] glass-card rounded-2xl overflow-hidden shadow-2xl animate-scale-in flex flex-col">
        
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-slate-700/50 bg-slate-900/50">
          <h2 className="text-xl font-bold text-slate-100 pr-8">
            {summary.title}
          </h2>
          <button
            onClick={onClose}
            className="p-2 rounded-full hover:bg-slate-700/50 transition-colors"
          >
            <X className="h-6 w-6 text-slate-400" />
          </button>
        </div>

        {/* Scrollable Body */}
        <div className="p-6 overflow-y-auto custom-scrollbar">
          
          {/* Original Summary (Short) */}
          <div className="bg-slate-800/30 rounded-xl p-4">
            <h3 className="text-sm font-semibold text-slate-400 mb-2">概要</h3>
            <p className="text-sm text-slate-400">
              {summary.summary}
            </p>
          </div>

          <div className="my-8 border-t border-slate-700/50" />

          {/* Detailed Summary */}
          <div className="prose prose-invert max-w-none">
            <h3 className="text-lg font-semibold text-accent-300 mb-4">詳細要約</h3>
            <div className="prose prose-invert prose-sm max-w-none text-slate-300">
              <ReactMarkdown remarkPlugins={[remarkBreaks]}>
                {summary.detailSummary || "詳細な要約は現在作成中です。"}
              </ReactMarkdown>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-slate-700/50 bg-slate-900/50 text-right">
          <button
            onClick={onClose}
            className="px-6 py-2 rounded-lg bg-primary-600 hover:bg-primary-500 text-white font-medium transition-colors"
          >
            閉じる
          </button>
        </div>
      </div>
    </div>
  );
}

export default SummaryModal;
