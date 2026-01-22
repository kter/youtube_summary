import { useState } from 'react';
import { Calendar, ExternalLink, FileText } from 'lucide-react';
import NewBadge from './NewBadge';
import SummaryModal from './SummaryModal';

function VideoCard({ summary, style }) {
  const [isModalOpen, setIsModalOpen] = useState(false);

  const {
    videoId,
    title,
    channelTitle,
    publishedAt,
    thumbnails,
    summary: summaryText,
    processedAt,
  } = summary;

  // Get thumbnail URL (prefer medium, fallback to default)
  const thumbnailUrl = thumbnails?.medium?.url || thumbnails?.default?.url || 
    `https://i.ytimg.com/vi/${videoId}/mqdefault.jpg`;

  // Format date
  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ja-JP', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const youtubeUrl = `https://www.youtube.com/watch?v=${videoId}`;

  return (
    <>
      <article
        className="glass-card rounded-2xl overflow-hidden animate-fade-in group hover:shadow-2xl hover:shadow-primary-500/10 transition-all duration-300"
        style={style}
        onClick={() => setIsModalOpen(true)}
      >
        {/* Thumbnail */}
        <div className="relative cursor-pointer overflow-hidden">
          <img
            src={thumbnailUrl}
            alt={title}
            className="w-full aspect-video object-cover transition-transform duration-500 group-hover:scale-105"
            loading="lazy"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-60" />
          
          {/* External Link Button Overlay */}
          <div className="absolute bottom-3 right-3 z-10">
            <a
              href={youtubeUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="p-2 rounded-full bg-white/10 hover:bg-white/20 backdrop-blur-md transition-colors text-white"
              onClick={(e) => e.stopPropagation()}
            >
              <ExternalLink className="h-5 w-5" />
            </a>
          </div>
          
          {/* NEW Badge */}
          <NewBadge processedAt={processedAt} />
        </div>

        {/* Content */}
        <div className="p-5 cursor-pointer">
          {/* Title */}
          <h3 className="font-semibold text-slate-100 line-clamp-2 mb-2 group-hover:text-primary-400 transition-colors">
            {title}
          </h3>

          {/* Channel */}
          <p className="text-sm text-slate-400 mb-3">{channelTitle}</p>

          {/* Stats */}
          <div className="flex items-center justify-between text-sm text-slate-500 mb-4">
            <div className="flex items-center gap-1">
              <Calendar className="h-4 w-4" />
              <span>{formatDate(publishedAt)}</span>
            </div>
            
            <div className="flex items-center gap-1 text-primary-400 font-medium opacity-0 group-hover:opacity-100 transition-opacity">
              <FileText className="h-4 w-4" />
              <span>詳細を読む</span>
            </div>
          </div>

          {/* Summary */}
          <div className="pt-4 border-t border-slate-700/50">
            <p className="text-sm text-slate-300 leading-relaxed line-clamp-3">
              {summaryText}
            </p>
          </div>
        </div>
      </article>

      <SummaryModal 
        summary={summary} 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
      />
    </>
  );
}

export default VideoCard;
